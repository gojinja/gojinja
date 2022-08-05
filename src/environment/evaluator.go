package environment

import (
	"fmt"
	"github.com/gojinja/gojinja/src/lexer"
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/operator"
)

type evaluator struct {
	ctx      *renderContext
	renderer *renderer
	frame    *frame
}

func evalExpr(ctx *renderContext, frame *frame, expr nodes.Expr) (any, error) {
	evaluator := &evaluator{
		ctx:   ctx,
		frame: frame,
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
		return ev.evalCompare(e)
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

func (ev *evaluator) evalBinExprOnAny(left, right any, op string) (any, error) {
	switch op {
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
		// Membership
	case "in":
		return operator.Contains(right, left)
	case "notin":
		{
			b, err := operator.Contains(right, left)
			return !b, err
		}
	default:
		panic(fmt.Sprintf("unexpected operator `%v`", op))
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
	return ev.evalBinExprOnAny(left, right, expr.Op)
}

func (ev *evaluator) evalCompare(expr *nodes.Compare) (any, error) {
	// TODO this will probably need to be reimplemented, because python allows chaining of comparisons
	// 5 == (1+4) == (3+2)
	// should evaluate to true
	acc, err := ev.evalExpr(expr.Expr)
	if err != nil {
		return nil, err
	}

	for _, op := range expr.Ops {
		right, err := ev.evalExpr(op.Expr)
		if err != nil {
			return nil, err
		}
		acc, err = ev.evalBinExprOnAny(acc, right, op.Op)
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func (ev *evaluator) evalName(name *nodes.Name) (any, error) {
	//ref = frame.symbols.ref(node.name)

	if name.Ctx == "store" && (ev.frame.toplevel || ev.frame.loopFrame || ev.frame.blockFrame) {
		// TODO support stores in expressions
		//if ev.frame.assignStack != nil {
		//	ev.frame.assignStack.lastElement().add(name.Name)
		//}
	}

	ref, err := ev.frame.symbols.ref(name.Name)
	if err != nil {
		return nil, err
	}

	// If we are looking up a variable we might have to deal with the
	// case where it's undefined.  We can skip that case if the load
	// instruction indicates a parameter which are always defined.
	if name.Ctx == "load" {
		load := ev.frame.symbols.findLoad(ref)
		if !(load != nil && load.variant == varLoadParameter && !ev.renderer.parameterIsUndeclared(ref)) {
			//if {ref} is missing {
			//	return undefined(name={node.name!r}), nil
			//}
			//return ref, nil
			panic("this case is not implemented")
		}
	}
	return ref, nil
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

func coerceBool(ctx *renderContext, frame *frame, expr nodes.Expr) (bool, error) {
	cond, err := evalExpr(ctx, frame, expr)
	if err != nil {
		return false, err
	}
	return operator.Bool(cond)
}
