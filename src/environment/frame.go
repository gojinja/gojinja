package environment

import (
	"fmt"
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/utils/iterator"
	"github.com/gojinja/gojinja/src/utils/set"
	"golang.org/x/exp/maps"
	"reflect"
)

const (
	varLoadParameter = "param"
	varLoadResolve   = "resolve"
	varLoadAlias     = "alias"
	varLoadUndefined = "undefined"
)

type symbolLoad struct {
	variant string  // One of the constants above
	target  *string // The name of the symbol (maybe nil)
}

type symbols struct {
	level  int
	parent *symbols
	refs   map[string]string
	loads  map[string]symbolLoad
	stores set.Set[string]
}

func newSymbols(parent *symbols, level *int) *symbols {
	var lvl int
	if level != nil {
		lvl = *level
	} else {
		if parent == nil {
			lvl = 0
		} else {
			lvl = parent.level + 1
		}
	}

	return &symbols{
		level:  lvl,
		parent: parent,
		refs:   make(map[string]string),
		loads:  make(map[string]symbolLoad),
		stores: set.New[string](),
	}
}

func (s *symbols) declareParameter(name string) string {
	s.stores.Add(name)
	return s.defineRef(name, &symbolLoad{variant: varLoadParameter, target: nil})
}

func (s *symbols) store(name string) {
	s.stores.Add(name)

	// If we have not seen the name referenced yet, we need to figure
	// out what to set it to.

	if _, ok := s.refs[name]; !ok {
		// If there is a parent scope we check if the name has a
		// reference there.  If it does it means we might have to alias
		// to a variable there.

		if s.parent != nil {
			outerRef := s.parent.findRef(name)
			if outerRef != nil {
				s.defineRef(name, &symbolLoad{variant: varLoadAlias, target: outerRef})
				return
			}
		}

		// Otherwise we can just set it to undefined.
		s.defineRef(name, &symbolLoad{variant: varLoadUndefined, target: nil})
	}
}

func (s *symbols) defineRef(name string, symbol *symbolLoad) string {
	ident := fmt.Sprintf("l_%d_%s", s.level, name)
	s.refs[name] = ident
	if symbol != nil {
		s.loads[ident] = *symbol
	}
	return ident
}

func (s *symbols) load(name string) {
	if s.findRef(name) == nil {
		s.defineRef(name, &symbolLoad{variant: varLoadResolve, target: &name})
	}
}

func (s *symbols) branchUpdate(branchSymbols []*symbols) {
	stores := make(map[string]int)

	for _, branch := range branchSymbols {
		for target := range branch.stores {
			if s.stores.Has(target) {
				continue
			}
			stores[target] += 1
		}
	}

	for _, branch := range branchSymbols {
		maps.Copy(s.refs, branch.refs)
		maps.Copy(s.loads, branch.loads)
		maps.Copy(s.stores, branch.stores)
	}

	for name, branchCount := range stores {
		if branchCount == len(branchSymbols) {
			continue
		}

		target := s.findRef(name)
		if target == nil {
			panic("target shouldn't be nil (it's a bug in jinja)")
		}

		if s.parent != nil {
			outerTarget := s.parent.findRef(name)
			if outerTarget != nil {
				s.loads[*target] = symbolLoad{variant: varLoadAlias, target: outerTarget}
				continue
			}
		}
		targetName := name
		s.loads[*target] = symbolLoad{variant: varLoadResolve, target: &targetName}
	}
}

func (s *symbols) copy() *symbols {
	return &symbols{
		level:  s.level,
		parent: s.parent,
		refs:   maps.Clone(s.refs),
		loads:  maps.Clone(s.loads),
		stores: s.stores.Clone(),
	}
}

func (s *symbols) findLoad(target string) *symbolLoad {
	if v, ok := s.loads[target]; ok {
		return &v
	}
	if s.parent != nil {
		return s.parent.findLoad(target)
	}
	return nil
}

func (s *symbols) findRef(name string) *string {
	if v, ok := s.refs[name]; ok {
		return &v
	}
	if s.parent != nil {
		return s.parent.findRef(name)
	}
	return nil
}

func (s *symbols) ref(name string) (string, error) {
	rv := s.findRef(name)
	if rv == nil {
		return "", fmt.Errorf("tried to resolve a name to a reference that was unknown to the frame %q", name)
	}
	return *rv, nil
}

func (s *symbols) analyzeNode(node nodes.Node, args map[string]any) error {
	return newRootVisitor(s).visit(node, args)
}

type frame struct {
	evalCtx            *evalContext
	parent             *frame
	symbols            *symbols
	requireOutputCheck bool
	buffer             *string
	block              *string // the name of the block we're in, otherwise nil
	toplevel           bool    // a toplevel frame is the root + soft frames such as if conditions.
	// the root frame is basically just the outermost frame, so no if
	// conditions. This information is used to optimize inheritance situations.
	rootlevel bool
	// variables set inside of loops and blocks should not affect outer frames,
	// but they still need to be kept track of as part of the active context.
	loopFrame  bool
	blockFrame bool
	// track whether the frame is being used in an if-statement or conditional
	// expression as it determines which errors should be raised during runtime
	// or compile time.
	softFrame bool
}

func newFrame(evalCtx *evalContext, parent *frame, level *int) *frame {
	f := &frame{
		evalCtx: evalCtx,
		parent:  parent,
	}

	if parent == nil {
		f.symbols = newSymbols(nil, level)
	} else {
		f.symbols = newSymbols(parent.symbols, level)
		f.requireOutputCheck = parent.requireOutputCheck
		f.buffer = parent.buffer
		f.block = parent.block
	}

	return f
}

func (f *frame) copy() *frame {
	return &frame{
		evalCtx:            f.evalCtx,
		parent:             f.parent,
		symbols:            f.symbols.copy(),
		requireOutputCheck: f.requireOutputCheck,
		buffer:             f.buffer,
		block:              f.block,
		toplevel:           f.toplevel,
		rootlevel:          f.rootlevel,
		loopFrame:          f.loopFrame,
		blockFrame:         f.blockFrame,
		softFrame:          f.softFrame,
	}
}

func (f *frame) soft() *frame {
	rv := f.copy()
	rv.rootlevel = false
	rv.softFrame = true
	return rv
}

type frameSymbolVisitor struct {
	symbols *symbols
}

func (f *frameSymbolVisitor) genericVisit(node nodes.Node, args map[string]any) {
	children := node.IterChildNodes(nil, nil)
	for children.HasNext() {
		child, _ := children.Next()
		f.visit(child, args)
	}
}

func (f *frameSymbolVisitor) visit(node nodes.Node, args map[string]any) {
	switch v := node.(type) {
	case *nodes.Name:
		f.visitName(v, args)
	case *nodes.NSRef:
		f.symbols.load(v.Name)
	case *nodes.If:
		f.visitIf(v, args)
	case *nodes.Macro:
		f.symbols.store(v.Name)
	case *nodes.Import:
		f.genericVisit(node, args)
		f.symbols.store(v.Target)
	case *nodes.FromImport:
		f.genericVisit(node, args)
		for _, name := range v.Names {
			if len(name) > 1 {
				f.symbols.store(name[1])
			} else {
				f.symbols.store(name[0])
			}
		}
	case *nodes.Assign:
		f.visit(v.Node, args)
		f.visit(v.Target, args)
	case *nodes.For:
		f.visit(v.Iter, args)
	case *nodes.CallBlock:
		f.visit(v.Call, args)
	case *nodes.FilterBlock:
		f.visit(v.Filter, args)
	case *nodes.With:
		for _, target := range v.Values {
			f.visit(target, nil)
		}
	case *nodes.AssignBlock:
		f.visit(v.Target, args)
	case *nodes.Scope:
		// """Stop visiting at scopes."""
	case *nodes.Block:
		// """Stop visiting at blocks."""
	//case *nodes.OverlayScope:
	//	// """Do not visit into overlay scopes."""

	default:
		f.genericVisit(node, args)
	}
}

func (f *frameSymbolVisitor) visitName(node *nodes.Name, args map[string]any) {
	storeAsParamAny, ok := args["store_as_param"]
	var storeAsParam bool
	if ok {
		storeAsParam, ok = storeAsParamAny.(bool)
	}

	if ok && storeAsParam || node.Ctx == "param" {
		f.symbols.declareParameter(node.Name)
	} else if node.Ctx == "store" {
		f.symbols.store(node.Name)
	} else if node.Ctx == "load" {
		f.symbols.load(node.Name)
	}
}

func (f *frameSymbolVisitor) visitIf(node *nodes.If, args map[string]any) {
	f.visit(node.Test, args)
	originalSymbols := f.symbols

	innerVisit := func(nodes []nodes.Node) *symbols {
		f.symbols = originalSymbols.copy()
		rv := f.symbols

		for _, subnode := range nodes {
			f.visit(subnode, args)
		}

		f.symbols = originalSymbols
		return rv
	}

	bodySymbols := innerVisit(node.Body)
	elifs, _ := iterator.ToSlice(iterator.Map(iterator.FromSlice(node.Elif), func(elif *nodes.If) (nodes.Node, error) {
		return elif, nil
	}))
	elifSymbols := innerVisit(elifs)
	elseSymbols := innerVisit(node.Else)
	f.symbols.branchUpdate([]*symbols{bodySymbols, elifSymbols, elseSymbols})
}

type rootVisitor struct {
	symbolVisitor *frameSymbolVisitor
}

func newRootVisitor(s *symbols) *rootVisitor {
	return &rootVisitor{
		symbolVisitor: &frameSymbolVisitor{
			symbols: s,
		},
	}
}

func (r *rootVisitor) simpleVisit(node nodes.Node, args map[string]any) {
	for children := node.IterChildNodes(nil, nil); children.HasNext(); {
		child, _ := children.Next()
		r.symbolVisitor.visit(child, nil)
	}
}

func (r *rootVisitor) genericVisit(node nodes.Node, args map[string]any) error {
	return fmt.Errorf("not implemented error: cannot find symbols for %s", reflect.TypeOf(node))
}

func (r *rootVisitor) visit(node nodes.Node, args map[string]any) error {
	switch v := node.(type) {
	case *nodes.AssignBlock:
		for _, child := range v.Body {
			r.symbolVisitor.visit(child, args)
		}
	case *nodes.CallBlock:
		children := node.IterChildNodes([]string{"call"}, nil)
		for children.HasNext() {
			child, _ := children.Next()
			r.symbolVisitor.visit(child, nil)
		}
	//case *nodes.OverlayScope:  // TODO add nodes.OverlayScope
	//	for _, child := range v.Body {
	//		r.symbolVisitor.visit(child, nil)
	//	}
	case *nodes.For:
		return r.visitFor(v, args)

	case *nodes.With:
		for _, target := range v.Targets {
			r.symbolVisitor.visit(target, nil)
		}
		for _, child := range v.Body {
			r.symbolVisitor.visit(child, nil)
		}

		// The rest of explicitly defined visitors just call simpleVisit
	case *nodes.Template, *nodes.Block, *nodes.Macro, *nodes.FilterBlock, *nodes.Scope, *nodes.If, *nodes.ScopedEvalContextModifier:
		r.simpleVisit(node, args)

	default:
		return r.genericVisit(node, args)
	}
	return nil
}

func (r *rootVisitor) visitFor(v *nodes.For, args map[string]any) error {
	forBranchAny, ok := args["for_branch"]
	var forBranch string
	if ok {
		forBranch, ok = forBranchAny.(string)
	}
	if !ok {
		// TODO it's possible that no arg is acceptable and default "body" value should be used
		return fmt.Errorf("missing or invalid for_branch argument (should be passed to analyzeNode)")
	}

	var branch []nodes.Node
	if forBranch == "body" {
		r.symbolVisitor.visit(v.Target, map[string]any{"store_as_param": true})
		branch = v.Body
	} else if forBranch == "else" {
		branch = v.Else
	} else if forBranch == "test" {
		r.symbolVisitor.visit(v.Target, map[string]any{"store_as_param": true})
		if v.Test != nil {
			r.symbolVisitor.visit(*v.Test, nil)
		}
		return nil
	} else {
		return fmt.Errorf("invalid for_branch argument value %q", forBranch)
	}
	if branch != nil {
		for _, child := range branch {
			r.symbolVisitor.visit(child, nil)
		}
	}

	return nil
}
