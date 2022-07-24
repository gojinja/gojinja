package nodes

import "golang.org/x/exp/slices"

type Node interface {
	GetLineno() int
	SetCtx(ctx string)
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

func (n *NodeCommon) GetLineno() int {
	return n.Lineno
}

type ExprCommon NodeCommon

func (e *ExprCommon) GetLineno() int {
	return e.Lineno
}

type StmtCommon NodeCommon

func (s *StmtCommon) GetLineno() int {
	return s.Lineno
}

func (ExprCommon) CanAssign() bool {
	return false
}

type Template struct {
	Body []Node
	NodeCommon
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

type Output struct {
	Nodes []Expr
	StmtCommon
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

func (e *Extends) SetCtx(ctx string) {
	e.Template.SetCtx(ctx)
}

type MacroCall struct {
	Args     []Name
	Defaults []Expr
}

type Macro struct {
	Name string
	Body []Node
	MacroCall
	StmtCommon
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
	Options []Keyword
	StmtCommon
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

func (s *ScopedEvalContextModifier) SetCtx(ctx string) {
	for _, n := range s.Body {
		n.SetCtx(ctx)
	}
}

type Scope struct {
	Body []Node
	StmtCommon
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

func (f *FilterBlock) SetCtx(ctx string) {
	for _, n := range f.Body {
		n.SetCtx(ctx)
	}
	f.Filter.SetCtx(ctx)
}

type Literal Expr
type LiteralCommon ExprCommon

func (l LiteralCommon) GetLineno() int {
	return l.Lineno
}

func (LiteralCommon) CanAssign() bool {
	return false
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

func (n Name) CanAssign() bool {
	return !slices.Contains([]string{"true", "false", "none", "True", "False", "None"}, n.Name)
}

func (n Name) GetName() string {
	return n.Name
}

type NSRef struct {
	Name string
	Attr string
	ExprCommon
}

func (n NSRef) SetCtx(string) {}

func (n NSRef) CanAssign() bool {
	return true
}

func (n NSRef) GetName() string {
	return n.Name
}

type CondExpr struct {
	Test  Expr
	Expr1 Expr
	Expr2 *Expr
	ExprCommon
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

func (h HelperCommon) GetLineno() int {
	return h.Lineno
}

type Operand struct {
	Op   string
	Expr Node
	HelperCommon
}

func (o *Operand) SetCtx(ctx string) {
	o.Expr.SetCtx(ctx)
}

type Compare struct {
	Expr Expr
	Ops  []Operand
	ExprCommon
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

func (b *BinExpr) SetCtx(ctx string) {
	b.Left.SetCtx(ctx)
	b.Right.SetCtx(ctx)
}

type Concat struct {
	Nodes []Expr
	ExprCommon
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

func (u *UnaryExpr) SetCtx(ctx string) {
	u.Node.SetCtx(ctx)
}

type Getattr struct {
	Node Expr
	Attr string
	Ctx  string
	ExprCommon
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

func (g *Getitem) SetCtx(ctx string) {
	g.Node.SetCtx(ctx)
}

type Slice struct {
	Start *Expr
	Stop  *Expr
	Step  *Expr
	ExprCommon
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
	Kwargs    []Keyword
	DynArgs   *Expr
	DynKwargs *Expr
	ExprCommon
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

type Import struct {
	Template    Expr
	WithContext bool
	Target      string
	StmtCommon
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
	Kwargs    []Keyword // Jinja uses Pair but then other methods returns Keyword instead of Pair -_-
	DynArgs   *Expr
	DynKwargs *Expr
	ExprCommon
}

type Filter struct {
	FilterTestCommon
}

func (f *Filter) SetCtx(ctx string) {
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

func (k *Keyword) SetCtx(ctx string) {
	k.Value.SetCtx(ctx)
}

type If struct {
	Test Node
	Body []Node
	Elif []If
	Else []Node
	StmtCommon
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
	Call Call
	Body []Node
	MacroCall
	StmtCommon
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
var _ Stmt = &Output{}
var _ Stmt = &CallBlock{}
var _ Stmt = &Include{}
var _ Stmt = &Import{}
var _ Stmt = &Assign{}
var _ Stmt = &AssignBlock{}
var _ Stmt = &With{}
var _ Stmt = &For{}

var _ SetWithContexter = &Include{}
var _ SetWithContexter = &Import{}

var _ Expr = &BinExpr{}
var _ Expr = &UnaryExpr{}
var _ Expr = &CondExpr{}
var _ Expr = &Compare{}
var _ Expr = &Concat{}
var _ Expr = &Call{}
var _ Expr = &Filter{}
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

var _ Helper = &Keyword{}
var _ Helper = &Operand{}
