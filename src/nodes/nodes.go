package nodes

import (
	"github.com/gojinja/gojinja/src/utils/iterator"
	"golang.org/x/exp/slices"
)

// TODO check if all Nodes have fields in the same order as in jinja (and yield them in this order in IterChildNodes)

type Node interface {
	GetLineno() int
	SetCtx(ctx string)
	IterChildNodes(exclude, only []string) iterator.Iterator[Node]
}

func asNode[T Node](v T) (Node, error) {
	return v, nil
}

type ExprWithName interface {
	GetName() string
	Expr
}

type SetWithContexter interface {
	SetWithContext(bool)
	Stmt
}

type NodeCommon struct {
	Lineno int
}

func (*NodeCommon) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	return iterator.Empty[Node]()
}

func (n *NodeCommon) GetLineno() int {
	return n.Lineno
}

type ExprCommon NodeCommon

func (*ExprCommon) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	return iterator.Empty[Node]()
}

func (e *ExprCommon) GetLineno() int {
	return e.Lineno
}

type StmtCommon NodeCommon

func (*StmtCommon) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	return iterator.Empty[Node]()
}

func (s *StmtCommon) GetLineno() int {
	return s.Lineno
}

type StmtWithNodes interface {
	GetNodes() []Expr
	Stmt
}

func (*ExprCommon) CanAssign() bool {
	return false
}

type Template struct {
	Body []Node
	NodeCommon
}

func includeField(name string, exclude, only []string) bool {
	// Check for nil is redundant, it's a microoptimization - it'll almost always be nil.
	if exclude != nil && slices.Contains(exclude, name) {
		return false
	}
	if only != nil && !slices.Contains(only, name) {
		return false
	}
	return true
}

func (t *Template) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("body", exclude, only) {
		return iterator.FromSlice(t.Body)
	}
	return iterator.Empty[Node]()
}

func (t *Template) SetCtx(ctx string) {
	for _, n := range t.Body {
		n.SetCtx(ctx)
	}
}

type Stmt interface {
	Node
}

type Expr interface {
	CanAssign() bool
	Node
}

type Block struct {
	Name     string
	Body     []Node
	Scoped   bool
	Required bool
	StmtCommon
}

func (b *Block) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("body", exclude, only) {
		return iterator.FromSlice(b.Body)
	}
	return iterator.Empty[Node]()
}

func (b *Block) SetCtx(ctx string) {
	for _, n := range b.Body {
		n.SetCtx(ctx)
	}
}

type Output struct {
	Nodes []Expr
	StmtCommon
}

func (o *Output) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("nodes", exclude, only) {
		return iterator.Map(iterator.FromSlice(o.Nodes), asNode[Expr])
	}
	return iterator.Empty[Node]()
}

func (o *Output) GetNodes() []Expr {
	return o.Nodes
}

func (o *Output) SetCtx(ctx string) {
	for _, n := range o.Nodes {
		n.SetCtx(ctx)
	}
}

type Extends struct {
	Template Expr
	StmtCommon
}

func (e *Extends) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("template", exclude, only) {
		return iterator.Once(Node(e.Template))
	}
	return iterator.Empty[Node]()
}

func (e *Extends) SetCtx(ctx string) {
	e.Template.SetCtx(ctx)
}

type MacroCall struct {
	Args     []*Name
	Defaults []Expr
}

func (m *MacroCall) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("args", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(m.Args), asNode[*Name]))
	}
	if includeField("defaults", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(m.Defaults), asNode[Expr]))
	}
	return it
}

type Macro struct {
	Name string
	Body []Node
	MacroCall
	StmtCommon
}

func (m *Macro) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("body", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(m.Body))
	}
	return iterator.Chain(it, m.MacroCall.IterChildNodes(exclude, only))
}

func (m *Macro) SetCtx(ctx string) {
	for _, n := range m.Body {
		n.SetCtx(ctx)
	}
	for _, n := range m.Args {
		n.SetCtx(ctx)
	}
	for _, n := range m.Defaults {
		n.SetCtx(ctx)
	}
}

type EvalContextModifier struct {
	Options []*Keyword
	StmtCommon
}

func (e *EvalContextModifier) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("options", exclude, only) {
		return iterator.Map(iterator.FromSlice(e.Options), asNode[*Keyword])
	}
	return iterator.Empty[Node]()
}

func (e *EvalContextModifier) SetCtx(ctx string) {
	for _, n := range e.Options {
		n.SetCtx(ctx)
	}
}

type ScopedEvalContextModifier struct {
	Body []Node
	EvalContextModifier
}

func (s *ScopedEvalContextModifier) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("body", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(s.Body))
	}
	return iterator.Chain(it, s.EvalContextModifier.IterChildNodes(exclude, only))
}

func (s *ScopedEvalContextModifier) SetCtx(ctx string) {
	for _, n := range s.Body {
		n.SetCtx(ctx)
	}
}

type Scope struct {
	Body []Node
	StmtCommon
}

func (s *Scope) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("body", exclude, only) {
		return iterator.FromSlice(s.Body)
	}
	return iterator.Empty[Node]()
}

func (s *Scope) SetCtx(ctx string) {
	for _, n := range s.Body {
		n.SetCtx(ctx)
	}
}

type FilterBlock struct {
	Body   []Node
	Filter *Filter
	StmtCommon
}

func (f *FilterBlock) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("body", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(f.Body))
	}
	if includeField("filter", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(f.Filter)))
	}
	return it
}

func (f *FilterBlock) SetCtx(ctx string) {
	for _, n := range f.Body {
		n.SetCtx(ctx)
	}
	f.Filter.SetCtx(ctx)
}

type Literal Expr
type LiteralCommon ExprCommon

func (*LiteralCommon) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	return iterator.Empty[Node]()
}

func (l *LiteralCommon) GetLineno() int {
	return l.Lineno
}

func (*LiteralCommon) CanAssign() bool {
	return false
}

type List struct {
	Items []Expr
	LiteralCommon
}

func (l *List) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("items", exclude, only) {
		return iterator.Map(iterator.FromSlice(l.Items), asNode[Expr])
	}
	return iterator.Empty[Node]()
}

func (l *List) SetCtx(ctx string) {
	for _, i := range l.Items {
		i.SetCtx(ctx)
	}
}

type Pair struct {
	Key   Expr
	Value Expr
	HelperCommon
}

func (p *Pair) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("key", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(p.Key)))
	}
	if includeField("value", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(p.Value)))
	}
	return it
}

func (p *Pair) SetCtx(ctx string) {
	p.Key.SetCtx(ctx)
	p.Value.SetCtx(ctx)
}

type Dict struct {
	Items []*Pair
	LiteralCommon
}

func (d *Dict) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("items", exclude, only) {
		return iterator.Map(iterator.FromSlice(d.Items), asNode[*Pair])
	}
	return iterator.Empty[Node]()
}

func (d *Dict) SetCtx(ctx string) {
	for _, i := range d.Items {
		i.SetCtx(ctx)
	}
}

// TemplateData represents a constant template string.
type TemplateData struct {
	Data string
	LiteralCommon
}

func (t *TemplateData) SetCtx(string) {}

type Tuple struct {
	Items []Expr
	Ctx   string
	LiteralCommon
}

func (t *Tuple) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("items", exclude, only) {
		return iterator.Map(iterator.FromSlice(t.Items), asNode[Expr])
	}
	return iterator.Empty[Node]()
}

func (t *Tuple) SetCtx(ctx string) {
	for _, n := range t.Items {
		n.SetCtx(ctx)
	}
	t.Ctx = ctx
}

type Const struct {
	Value any
	LiteralCommon
}

func (c *Const) SetCtx(string) {}

type Name struct {
	Name string
	Ctx  string
	ExprCommon
}

func (n *Name) SetCtx(ctx string) {
	n.Ctx = ctx
}

func (n *Name) CanAssign() bool {
	return !slices.Contains([]string{"true", "false", "none", "True", "False", "None"}, n.Name)
}

func (n *Name) GetName() string {
	return n.Name
}

type NSRef struct {
	Name string
	Attr string
	ExprCommon
}

func (n *NSRef) SetCtx(string) {}

func (n *NSRef) CanAssign() bool {
	return true
}

func (n *NSRef) GetName() string {
	return n.Name
}

type CondExpr struct {
	Test  Expr
	Expr1 Expr
	Expr2 *Expr
	ExprCommon
}

func (c *CondExpr) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("test", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(c.Test)))
	}
	if includeField("expr1", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(c.Expr1)))
	}
	if c.Expr2 != nil && includeField("expr2", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*c.Expr2)))
	}
	return it
}

func (c *CondExpr) SetCtx(ctx string) {
	c.Test.SetCtx(ctx)
	c.Expr1.SetCtx(ctx)
	if c.Expr2 != nil {
		(*c.Expr2).SetCtx(ctx)
	}
}

type Helper Node
type HelperCommon NodeCommon

func (*HelperCommon) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	return iterator.Empty[Node]()
}

func (h *HelperCommon) GetLineno() int {
	return h.Lineno
}

type Operand struct {
	Op   string
	Expr Expr
	HelperCommon
}

func (o *Operand) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("expr", exclude, only) {
		return iterator.Once(Node(o.Expr))
	}
	return iterator.Empty[Node]()
}

func (o *Operand) SetCtx(ctx string) {
	o.Expr.SetCtx(ctx)
}

type Compare struct {
	Expr Expr
	Ops  []*Operand
	ExprCommon
}

func (c *Compare) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("expr", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(c.Expr)))
	}
	if includeField("ops", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(c.Ops), asNode[*Operand]))
	}
	return it
}

func (c *Compare) SetCtx(ctx string) {
	c.Expr.SetCtx(ctx)
	for _, n := range c.Ops {
		n.SetCtx(ctx)
	}
}

type BinExpr struct {
	Left  Expr
	Right Expr
	Op    string // same as lexer.TokenAdd etc. + "and", "or"
	ExprCommon
}

func (b *BinExpr) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("left", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(b.Left)))
	}
	if includeField("rif=ght", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(b.Right)))
	}
	return it
}

func (b *BinExpr) SetCtx(ctx string) {
	b.Left.SetCtx(ctx)
	b.Right.SetCtx(ctx)
}

type Concat struct {
	Nodes []Expr
	ExprCommon
}

func (c *Concat) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("nodes", exclude, only) {
		return iterator.Map(iterator.FromSlice(c.Nodes), asNode[Expr])
	}
	return iterator.Empty[Node]()
}

func (c *Concat) SetCtx(ctx string) {
	for _, n := range c.Nodes {
		n.SetCtx(ctx)
	}
}

type UnaryExpr struct {
	Node Expr
	Op   string // same as lexer.TokenAdd etc. + "not"
	ExprCommon
}

func (u *UnaryExpr) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("node", exclude, only) {
		return iterator.Once(Node(u.Node))
	}
	return iterator.Empty[Node]()
}

func (u *UnaryExpr) SetCtx(ctx string) {
	u.Node.SetCtx(ctx)
}

type Getattr struct {
	Node Expr
	Attr string
	Ctx  string
	ExprCommon
}

func (g *Getattr) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("node", exclude, only) {
		return iterator.Once(Node(g.Node))
	}
	return iterator.Empty[Node]()
}

func (g *Getattr) SetCtx(ctx string) {
	g.Node.SetCtx(ctx)
}

type Getitem struct {
	Node Expr
	Arg  Expr
	Ctx  string
	ExprCommon
}

func (g *Getitem) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("node", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(g.Node)))
	}
	if includeField("arg", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(g.Arg)))
	}
	return it
}

func (g *Getitem) SetCtx(ctx string) {
	g.Node.SetCtx(ctx)
}

type Slice struct {
	Start *Expr
	Stop  *Expr
	Step  *Expr
	ExprCommon
}

func (s *Slice) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if s.Start != nil && includeField("start", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*s.Start)))
	}
	if s.Stop != nil && includeField("stop", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*s.Stop)))
	}
	if s.Step != nil && includeField("step", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*s.Step)))
	}
	return it
}

func (s *Slice) SetCtx(ctx string) {
	if s.Start != nil {
		(*s.Start).SetCtx(ctx)
	}
	if s.Stop != nil {
		(*s.Stop).SetCtx(ctx)
	}
	if s.Step != nil {
		(*s.Step).SetCtx(ctx)
	}
}

type Call struct {
	Node      Expr
	Args      []Expr
	Kwargs    []*Keyword
	DynArgs   *Expr
	DynKwargs *Expr
	ExprCommon
}

func (c *Call) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("node", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(c.Node)))
	}
	if includeField("args", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(c.Args), asNode[Expr]))
	}
	if includeField("kwargs", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(c.Kwargs), asNode[*Keyword]))
	}
	if c.DynArgs != nil && includeField("dynargs", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*c.DynArgs)))
	}
	if c.DynKwargs != nil && includeField("dynkwargs", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*c.DynKwargs)))
	}
	return it
}

func (c *Call) SetCtx(ctx string) {
	c.Node.SetCtx(ctx)
	for _, n := range c.Args {
		n.SetCtx(ctx)
	}
	for _, n := range c.Kwargs {
		n.SetCtx(ctx)
	}
	if c.DynArgs != nil {
		(*c.DynArgs).SetCtx(ctx)
	}
	if c.DynKwargs != nil {
		(*c.DynKwargs).SetCtx(ctx)
	}
}

type Include struct {
	Template      Expr
	WithContext   bool
	IgnoreMissing bool
	StmtCommon
}

func (i *Include) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("template", exclude, only) {
		return iterator.Once(Node(i.Template))
	}
	return iterator.Empty[Node]()
}

func (i *Include) SetWithContext(b bool) {
	i.WithContext = b
}

func (i *Include) SetCtx(ctx string) {
	i.Template.SetCtx(ctx)
}

type Assign struct {
	Target Expr
	Node   Node
	StmtCommon
}

func (a *Assign) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("target", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(a.Target)))
	}
	if includeField("node", exclude, only) {
		it = iterator.Chain(it, iterator.Once(a.Node))
	}
	return it
}

func (a *Assign) SetCtx(ctx string) {
	a.Target.SetCtx(ctx)
	a.Node.SetCtx(ctx)
}

type AssignBlock struct {
	Target Expr
	Body   []Node
	Filter *Filter
	StmtCommon
}

func (a *AssignBlock) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("target", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(a.Target)))
	}
	if includeField("body", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(a.Body))
	}
	if a.Filter != nil && includeField("filter", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(a.Filter)))
	}
	return it
}

func (a *AssignBlock) SetCtx(ctx string) {
	a.Target.SetCtx(ctx)
	for _, n := range a.Body {
		n.SetCtx(ctx)
	}
	if a.Filter != nil {
		(*a.Filter).SetCtx(ctx)
	}
}

type With struct {
	Targets []Expr
	Values  []Expr
	Body    []Node
	StmtCommon
}

func (w *With) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("targets", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(w.Targets), asNode[Expr]))
	}
	if includeField("values", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(w.Values), asNode[Expr]))
	}
	if includeField("body", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(w.Body))
	}
	return it
}

func (w *With) SetCtx(ctx string) {
	for _, n := range w.Body {
		n.SetCtx(ctx)
	}
	for _, n := range w.Targets {
		n.SetCtx(ctx)
	}
	for _, n := range w.Values {
		n.SetCtx(ctx)
	}
}

type FromImport struct {
	Template    Expr
	WithContext bool
	Names       [][]string // name or name with alias
	StmtCommon
}

func (f *FromImport) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("template", exclude, only) {
		return iterator.Once(Node(f.Template))
	}
	return iterator.Empty[Node]()
}

func (f *FromImport) SetWithContext(b bool) {
	f.WithContext = b
}

func (f *FromImport) SetCtx(ctx string) {
	f.Template.SetCtx(ctx)
}

type Import struct {
	Template    Expr
	WithContext bool
	Target      string
	StmtCommon
}

func (i *Import) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("template", exclude, only) {
		return iterator.Once(Node(i.Template))
	}
	return iterator.Empty[Node]()
}

func (i *Import) SetWithContext(b bool) {
	i.WithContext = b
}

func (i *Import) SetCtx(ctx string) {
	i.Template.SetCtx(ctx)
}

type FilterTestCommon struct {
	Node      *Expr
	Name      string
	Args      []Expr
	Kwargs    []*Keyword // Jinja uses Pair but then other methods returns Keyword instead of Pair -_-
	DynArgs   *Expr
	DynKwargs *Expr
	ExprCommon
}

type Filter struct {
	FilterTestCommon
}

type Test struct {
	FilterTestCommon
}

func (f *FilterTestCommon) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if f.Node != nil && includeField("node", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*f.Node)))
	}
	if includeField("args", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(f.Args), asNode[Expr]))
	}
	if includeField("kwargs", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(f.Kwargs), asNode[*Keyword]))
	}
	if f.DynArgs != nil && includeField("dynargs", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*f.DynArgs)))
	}
	if f.DynKwargs != nil && includeField("dynkwargs", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(*f.DynKwargs)))
	}
	return it
}

func (f *FilterTestCommon) SetCtx(ctx string) {
	if f.Node != nil {
		(*f.Node).SetCtx(ctx)
	}
	for _, n := range f.Args {
		n.SetCtx(ctx)
	}
	for _, n := range f.Kwargs {
		n.SetCtx(ctx)
	}
	if f.DynArgs != nil {
		(*f.DynArgs).SetCtx(ctx)
	}
	if f.DynKwargs != nil {
		(*f.DynKwargs).SetCtx(ctx)
	}
}

type Keyword struct {
	Key   string
	Value Expr
	HelperCommon
}

func (k *Keyword) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	if includeField("value", exclude, only) {
		return iterator.Once(Node(k.Value))
	}
	return iterator.Empty[Node]()
}

func (k *Keyword) SetCtx(ctx string) {
	k.Value.SetCtx(ctx)
}

type If struct {
	Test Expr // jinja says it's Node, but having it as expr makes much more sense
	Body []Node
	Elif []*If
	Else []Node
	StmtCommon
}

func (i *If) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("test", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(i.Test)))
	}
	if includeField("body", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(i.Body))
	}
	if includeField("elif", exclude, only) {
		it = iterator.Chain(it, iterator.Map(iterator.FromSlice(i.Elif), asNode[*If]))
	}
	if includeField("else", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(i.Else))
	}
	return it
}

func (i *If) SetCtx(ctx string) {
	i.Test.SetCtx(ctx)
	for _, n := range i.Body {
		n.SetCtx(ctx)
	}
	for _, n := range i.Else {
		n.SetCtx(ctx)
	}
	for _, n := range i.Elif {
		n.SetCtx(ctx)
	}
}

type CallBlock struct {
	Call *Call
	Body []Node
	MacroCall
	StmtCommon
}

func (c *CallBlock) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if c.Call != nil && includeField("call", exclude, only) {
		it = iterator.Chain(it, iterator.Once(Node(c.Call)))
	}
	if includeField("body", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(c.Body))
	}
	return iterator.Chain(it, c.MacroCall.IterChildNodes(exclude, only))
}

func (c *CallBlock) SetCtx(ctx string) {
	c.Call.SetCtx(ctx)
	for _, n := range c.Args {
		n.SetCtx(ctx)
	}
	for _, n := range c.Defaults {
		n.SetCtx(ctx)
	}
	for _, n := range c.Body {
		n.SetCtx(ctx)
	}
}

type For struct {
	Target    Node
	Iter      Node
	Body      []Node
	Else      []Node
	Test      *Node
	Recursive bool
	StmtCommon
}

func (f *For) IterChildNodes(exclude, only []string) iterator.Iterator[Node] {
	it := iterator.Empty[Node]()
	if includeField("target", exclude, only) {
		it = iterator.Chain(it, iterator.Once(f.Target))
	}
	if includeField("iter", exclude, only) {
		it = iterator.Chain(it, iterator.Once(f.Iter))
	}
	if includeField("body", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(f.Body))
	}
	if includeField("else", exclude, only) {
		it = iterator.Chain(it, iterator.FromSlice(f.Else))
	}
	if f.Test != nil && includeField("test", exclude, only) {
		it = iterator.Chain(it, iterator.Once(*f.Test))
	}
	return it
}

func (f *For) SetCtx(ctx string) {
	f.Target.SetCtx(ctx)
	f.Iter.SetCtx(ctx)
	for _, n := range f.Body {
		n.SetCtx(ctx)
	}
	for _, n := range f.Else {
		n.SetCtx(ctx)
	}
	if f.Test != nil {
		(*f.Test).SetCtx(ctx)
	}
}

// Assert all types of nodes implement Node interface.
var _ Node = &Template{}

var _ Stmt = &Extends{}
var _ Stmt = &Macro{}
var _ Stmt = &Scope{}
var _ Stmt = &FilterBlock{}
var _ Stmt = &Output{}
var _ Stmt = &If{}
var _ Stmt = &ScopedEvalContextModifier{}
var _ Stmt = &EvalContextModifier{}
var _ Stmt = &CallBlock{}
var _ Stmt = &Include{}
var _ Stmt = &Import{}
var _ Stmt = &Assign{}
var _ Stmt = &AssignBlock{}
var _ Stmt = &With{}
var _ Stmt = &For{}
var _ Stmt = &Block{}
var _ Stmt = &FromImport{}

var _ StmtWithNodes = &Output{}

var _ SetWithContexter = &Include{}
var _ SetWithContexter = &Import{}
var _ SetWithContexter = &FromImport{}

var _ Expr = &BinExpr{}
var _ Expr = &UnaryExpr{}
var _ Expr = &CondExpr{}
var _ Expr = &Compare{}
var _ Expr = &Concat{}
var _ Expr = &Call{}
var _ Expr = &Filter{}
var _ Expr = &Test{}
var _ Expr = &Name{}
var _ Expr = &NSRef{}
var _ Expr = &Getattr{}
var _ Expr = &Getitem{}
var _ Expr = &Slice{}

var _ ExprWithName = &Name{}
var _ ExprWithName = &NSRef{}

var _ Literal = &Const{}
var _ Literal = &Tuple{}
var _ Literal = &TemplateData{}
var _ Literal = &List{}
var _ Literal = &Dict{}

var _ Helper = &Keyword{}
var _ Helper = &Operand{}
var _ Helper = &Pair{}
