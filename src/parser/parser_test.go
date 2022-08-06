package parser

import (
	"github.com/gojinja/gojinja/src/lexer"
	"github.com/gojinja/gojinja/src/nodes"
	"reflect"
	"testing"
)

type parserTest struct {
	input string
	res   *nodes.Template
}

var cases = []parserTest{
	{
		input: `{{ name }}`,
		res: &nodes.Template{
			Body: []nodes.Node{
				&nodes.Output{
					Nodes: []nodes.Expr{
						&nodes.Name{
							Name:       "name",
							Ctx:        "load",
							ExprCommon: nodes.ExprCommon{Lineno: 1},
						},
					},
					StmtCommon: nodes.StmtCommon{Lineno: 1},
				},
			},
			NodeCommon: nodes.NodeCommon{Lineno: 1},
		},
	},
	{
		input: `{% if abc != "OFF" %}my name is {{ abc }}{% endif %}{{ 5 + 1 }}`,
		res: &nodes.Template{
			Body: []nodes.Node{
				&nodes.If{
					Test: &nodes.Compare{
						Expr: &nodes.Name{
							Name:       "abc",
							Ctx:        "load",
							ExprCommon: nodes.ExprCommon{Lineno: 1},
						},
						Ops: []*nodes.Operand{
							{
								Op: "ne",
								Expr: &nodes.Const{
									Value:         "OFF",
									LiteralCommon: nodes.LiteralCommon{Lineno: 1},
								},
								HelperCommon: nodes.HelperCommon{Lineno: 1},
							},
						},
						ExprCommon: nodes.ExprCommon{Lineno: 1},
					},
					Body: []nodes.Node{
						&nodes.Output{
							Nodes: []nodes.Expr{
								&nodes.TemplateData{
									Data:          "my name is ",
									LiteralCommon: nodes.LiteralCommon{Lineno: 1},
								},
								&nodes.Name{
									Name:       "abc",
									Ctx:        "load",
									ExprCommon: nodes.ExprCommon{Lineno: 1},
								},
							},
							StmtCommon: nodes.StmtCommon{Lineno: 1},
						},
					},
					Elif:       []*nodes.If{},
					Else:       []nodes.Node{},
					StmtCommon: nodes.StmtCommon{Lineno: 1},
				},
				&nodes.Output{
					Nodes: []nodes.Expr{
						&nodes.BinExpr{
							Left: &nodes.Const{
								Value:         int64(5),
								LiteralCommon: nodes.LiteralCommon{Lineno: 1},
							},
							Right: &nodes.Const{
								Value:         int64(1),
								LiteralCommon: nodes.LiteralCommon{Lineno: 1},
							},
							Op:         lexer.TokenAdd,
							ExprCommon: nodes.ExprCommon{Lineno: 1},
						},
					},
					StmtCommon: nodes.StmtCommon{Lineno: 1},
				},
			},
			NodeCommon: nodes.NodeCommon{Lineno: 1},
		},
	},
}

func Test(t *testing.T) {
	for _, c := range cases {
		runCase(c, t)
	}
}

func getTokenStream(input string, t *testing.T) *lexer.TokenStream {
	info := lexer.DefaultEnvLexerInformation()
	l := lexer.GetLexer(info)
	s, err := l.Tokenize(input, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func runCase(c parserTest, t *testing.T) {
	ts := getTokenStream(c.input, t)
	p := NewParser(ts, nil, nil, nil, nil)
	template, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(template, c.res) {
		t.Fatalf("Expected %v, got %v", c.res, template)
	}
}
