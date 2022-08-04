package environment

import (
	"fmt"
	"github.com/gojinja/gojinja/src/lexer"
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/operator"
)

type evaluator struct {
	ctx *renderContext
}

func evalExpr(ctx *renderContext, expr nodes.Expr) (any, error) {
	evaluator := &evaluator{
		ctx: ctx,
	}
	v, err := evaluator.evalExpr(expr)
	if err != nil {
		return nil, fmt.Errorf("evaluation failure: %w", err)
	}
	return v, nil
}

func (ev *evaluator) evalExpr(expr nodes.Expr) (any, error) {
	// This could either be implemented in a polymorphic way (type-safe, because Eval() would be required by the interface)
	// or as a big-ass switch (see below).
	// Problems with polymorphism:
	// - nodes would have to be in the same package as context classes - that is in the environment package,
	// - poorer separation of concerns.
	// Problems with switch:
	// - switch can be non-exhaustive, it's the programmer's obligation to make sure all cases are covered

	switch e := expr.(type) {
	// Literals
	case *nodes.Const:
		return ev.evalLiteral(e)
	case *nodes.TemplateData:
		return ev.evalLiteral(e)
	case *nodes.Tuple:
		return ev.evalLiteral(e)
	case *nodes.List:
		return ev.evalLiteral(e)
	case *nodes.Dict:
		return ev.evalLiteral(e)
		// Others
	case *nodes.BinExpr:
		return ev.evalBinExpr(e)
	case *nodes.UnaryExpr:
		panic("unimplemented")
	case *nodes.CondExpr:
		panic("unimplemented")
	case *nodes.Compare:
		panic("unimplemented")
	case *nodes.Concat:
		panic("unimplemented")
	case *nodes.Call:
		panic("unimplemented")
	case *nodes.Filter:
		panic("unimplemented")
	case *nodes.Test:
		panic("unimplemented")
	case *nodes.Name:
		return ev.evalName(e)
	case *nodes.NSRef:
		panic("unimplemented")
	case *nodes.Getattr:
		panic("unimplemented")
	case *nodes.Getitem:
		panic("unimplemented")
	case *nodes.Slice:
		panic("unimplemented")
	default:
		panic(fmt.Sprintf("unexpected Expr type of value `%v` (this is a bug in gojinja)", expr))
	}
}

func (ev *evaluator) evalBinExpr(expr *nodes.BinExpr) (any, error) {
	left, err := ev.evalExpr(expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := ev.evalExpr(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Op {
	// Arithmetic
	case lexer.TokenAdd:
		return operator.Add(left, right)
	case lexer.TokenSub:
		return operator.Sub(left, right)
	case lexer.TokenMul:
		return operator.Mul(left, right)
	case lexer.TokenDiv:
		return operator.Div(left, right)
	case lexer.TokenFloordiv:
		return operator.FloorDiv(left, right)
	case lexer.TokenMod:
		return operator.Mod(left, right)
	case lexer.TokenPow:
		return operator.Pow(left, right)
		// Comparison
	case lexer.TokenEq:
		return operator.Eq(left, right)
	case lexer.TokenNe:
		return operator.Ne(left, right)
	case lexer.TokenGt:
		return operator.Gt(left, right)
	case lexer.TokenGteq:
		return operator.Ge(left, right)
	case lexer.TokenLt:
		return operator.Lt(left, right)
	case lexer.TokenLteq:
		return operator.Le(left, right)
	default:
		panic(fmt.Sprintf("unexpected operator `%v`", expr.Op))
	}
}

func (ev *evaluator) evalName(name *nodes.Name) (any, error) {
	return "EvaluatedNameValue", nil
	// TODO: implement
}

func (ev *evaluator) evalLiteral(lit nodes.Literal) (any, error) {
	switch v := lit.(type) {
	case *nodes.Const:
		return v.Value, nil
	case *nodes.TemplateData:
		return v.Data, nil
	case *nodes.Tuple:
		{
			tuple := make([]any, 0, len(v.Items))
			for _, itemExpr := range v.Items {
				item, err := ev.evalExpr(itemExpr)
				if err != nil {
					return nil, err
				}
				tuple = append(tuple, item)
			}
			return tuple, nil
		}
	case *nodes.List:
		{
			tuple := make([]any, 0, len(v.Items))
			for _, itemExpr := range v.Items {
				item, err := ev.evalExpr(itemExpr)
				if err != nil {
					return nil, err
				}
				tuple = append(tuple, item)
			}
			return tuple, nil
		}
	case *nodes.Dict:
		{
			dict := make(map[any]any, len(v.Items))
			for _, pair := range v.Items {
				key, err := ev.evalExpr(pair.Key)
				if err != nil {
					return nil, err
				}
				value, err := ev.evalExpr(pair.Value)
				if err != nil {
					return nil, err
				}
				dict[key] = value
			}
			return dict, nil
		}
	default:
		panic(fmt.Sprintf("unexpected Literal type of value `%v` (this is a bug in gojinja)", lit))
	}
}
