package environment

import (
	"fmt"
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/utils"
	"github.com/gojinja/gojinja/src/utils/iterator"
	"github.com/gojinja/gojinja/src/utils/set"
	"strings"
)

type output struct {
	builder strings.Builder
}

func (o *output) write(value any) error {
	// TODO expand as needed
	switch v := value.(type) {
	case string:
		if _, err := o.builder.WriteString(v); err != nil {
			return err
		}
	default:
		panic("unsupported type")
	}
	return nil
}

type renderer struct {
	out     *output
	ctx     *renderContext
	evalCtx *evalContext
	blocks  map[string]*nodes.Block
	// createdBlockContext bool
	// importAliases       map[string]string
	// extendsSoFar        int
	// codeLineno          int // TODO this is probably not needed
	hasKnownExtends bool
	//         # registry of all filters and tests (global, not block local)
	//        self.tests: t.Dict[str, str] = {}
	//        self.filters: t.Dict[str, str] = {}
	assignStack []set.Set[string]

	paramDefBlock []set.Set[string]
}

func renderTemplate(ctx *renderContext, node *nodes.Template) (iterator.Iterator[string], error) {
	// Renders the template piece by piece (iteration over statements).

	renderer := &renderer{
		evalCtx: ctx.evalCtx,
		ctx:     ctx,
		out:     &output{},
		blocks:  make(map[string]*nodes.Block),
	}

	blocks := nodes.FindAll[*nodes.Block](node)
	for blocks.HasNext() {
		block, _ := blocks.Next() // Error in walking over AST would be a bug in gojinja.
		if _, ok := renderer.blocks[block.Name]; ok {
			return nil, fmt.Errorf("block %q defined twice", block.Name)
		}
		renderer.blocks[block.Name] = block
	}

	// TODO process imports

	frame := newFrame(renderer.evalCtx, nil, nil)
	// TODO check for undeclared self & set if needed
	if err := frame.symbols.analyzeNode(node, nil); err != nil {
		return nil, fmt.Errorf("error preparing a frame for root node: %w", err)
	}
	frame.toplevel = true
	frame.rootlevel = true
	frame.requireOutputCheck = true // TODO jinja only sets it based on some conditions, but I believe we need it to be always true

	if err := renderer.validateAST(node); err != nil {
		// TODO shouldn't it be done before?
		return nil, err
	}

	// TODO lot's of stuff probably missing from here, but it's such a tangled mess...

	if err := renderer.enterFrame(frame); err != nil {
		return nil, err
	}
	// renderer.pull_dependencies(node.body)

	return renderer.renderTemplate(node, frame)
}

func (r *renderer) enterFrame(frame *frame) error {
	var undefs []string

	for target, symbolLoadInfo := range frame.symbols.loads {
		if symbolLoadInfo.variant == varLoadParameter {
			// do nothing
		} else if symbolLoadInfo.variant == varLoadResolve {
			// self.writeline(f"{param} = {self.get_resolve_func()}({param!r})")
			r.evalCtx.Set(target, r.ctx.ResolveOrMissing(*symbolLoadInfo.param)) // TODO potentially dangerous dereference?
		} else if symbolLoadInfo.variant == varLoadAlias {
			// self.writeline(f"{param} = {param}")
			v, ok := r.evalCtx.Get(*symbolLoadInfo.param) // TODO maybe ref wrap needed?
			if !ok {
				v = utils.GetMissing() // TODO correct?
			}
			r.evalCtx.Set(target, v)
		} else if symbolLoadInfo.variant == varLoadUndefined {
			undefs = append(undefs, target)
		} else {
			return fmt.Errorf("unknown load instruction")
		}
	}

	for _, target := range undefs {
		// TODO need to wrap param in ref?
		r.evalCtx.Set(target, utils.GetMissing())
	}
	return nil
}

func (r *renderer) validateAST(*nodes.Template) error {
	// TODO
	// This will probably only need to be done once, so maybe do it after parsing (not before rendering)?
	return nil
}

func (r *renderer) renderTemplate(node *nodes.Template, frame *frame) (iterator.Iterator[string], error) {
	return iterator.Map(iterator.FromSlice(node.Body), func(n nodes.Node) (string, error) {
		if err := r.renderNode(n, frame); err != nil {
			return "", err
		}
		s := r.out.builder.String()
		r.out.builder.Reset()
		return s, nil
	}), nil
}

func (r *renderer) renderNode(node nodes.Node, frame *frame) error {
	switch n := node.(type) {
	case *nodes.Extends:
		panic("not implemented")
	case *nodes.Macro:
		panic("not implemented")
	case *nodes.Scope:
		panic("not implemented")
	case *nodes.FilterBlock:
		panic("not implemented")
	case *nodes.Output:
		return r.renderOutput(n, frame)
	case *nodes.If:
		return r.renderIf(n, frame)
	case *nodes.ScopedEvalContextModifier:
		panic("not implemented")
	case *nodes.EvalContextModifier:
		panic("not implemented")
	case *nodes.CallBlock:
		panic("not implemented")
	case *nodes.Assign:
		panic("not implemented")
	case *nodes.AssignBlock:
		panic("not implemented")
	case *nodes.With:
		panic("not implemented")
	case *nodes.For:
		panic("not implemented")
	case *nodes.Block:
		panic("not implemented")
	// Include, Import, FromImport will probably require different handling (in renderTemplate)
	case *nodes.Include:
		panic("not implemented")
	case *nodes.Import:
		panic("not implemented")
	case *nodes.FromImport:
		panic("not implemented")
	default:
		// TODO it's possible that some of the nodes are missing in the switch, check this in the future
		panic(fmt.Sprintf("unexpected node type `%v` (this is a bug in gojinja)", node))
	}
}

func (r *renderer) blockvisit(body []nodes.Node, frame *frame) error {
	for _, n := range body {
		if err := r.renderNode(n, frame); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) renderIf(node *nodes.If, frame *frame) error {
	ifFrame := frame.soft()

	cond, err := coerceBool(r.ctx, frame, r, node.Test)
	if err != nil {
		return err
	}

	// if body
	if cond {
		return r.blockvisit(node.Body, ifFrame)
	}
	// else-ifs
	for _, elif := range node.Elif {
		cond, err := coerceBool(r.ctx, frame, r, node.Test)
		if err != nil {
			return err
		}
		if cond {
			return r.blockvisit(elif.Body, ifFrame)
		}
	}
	// else
	return r.blockvisit(node.Else, ifFrame)
}

func (r *renderer) renderOutput(node *nodes.Output, frame *frame) error {
	if frame.requireOutputCheck {
		if r.hasKnownExtends {
			return nil
		}
		// if !parentTemplate != nil {  // TODO
		// 	  return nil
		// }
	}

	for _, expr := range node.Nodes {
		if err := r.writeOutputChild(expr, frame); err != nil {
			return err
		}
	}

	return nil
}

func identity[T any](x T) T {
	return x
}

func (r *renderer) write(v any) error {
	// TODO add whatever support is needed for native mode
	return r.out.write(v)
}

func (r *renderer) writeOutputChild(child nodes.Expr, frame *frame) error {
	// TODO support finalize [substitute proper finalizer where identity is used]

	v, err := evalExpr(r.ctx, r, frame, child)
	if err != nil {
		return err
	}

	v = identity(v) // here <- finalizer

	// In default mode, `v` should be a string (after calling the finalizer).
	// In native mode, `v` can be any type.

	if !r.ctx.env.IsNative {
		if frame.evalCtx.volatile {
			if r.evalCtx.autoEscape {
				v = utils.Escape(v)
			} else {
				v = toString(v)
			}
		} else if frame.evalCtx.autoEscape {
			v = utils.Escape(v)
		} else {
			v = toString(v)
		}
	}

	return r.write(v)
}

func (r *renderer) parameterIsUndeclared(target string) bool {
	// TODO isn't it negated?
	if len(r.paramDefBlock) == 0 {
		return false
	}
	return r.paramDefBlock[len(r.paramDefBlock)-1].Has(target)
}
