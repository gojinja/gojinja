package operator

import (
	"fmt"
	"github.com/gojinja/gojinja/src/utils/slices"
	"math"
	"reflect"
	"strings"
)

func ToInt(v any) (i int64, ok bool) {
	if p, ok := v.(int); ok {
		return int64(p), ok
	}
	if p, ok := v.(int8); ok {
		return int64(p), ok
	}
	if p, ok := v.(int16); ok {
		return int64(p), ok
	}
	if p, ok := v.(int32); ok {
		return int64(p), ok
	}
	if p, ok := v.(uint); ok {
		return int64(p), ok
	}
	if p, ok := v.(uint8); ok {
		return int64(p), ok
	}
	if p, ok := v.(uint16); ok {
		return int64(p), ok
	}
	if p, ok := v.(uint32); ok {
		return int64(p), ok
	}
	if p, ok := v.(uint64); ok {
		return int64(p), ok
	}
	i, ok = v.(int64)
	return
}

func ToString(v any) string {
	return fmt.Sprint(v)
}

func ToFloat(v any) (f float64, ok bool) {
	if p, ok := v.(float32); ok {
		return float64(p), ok
	}
	f, ok = v.(float64)
	return
}

func ToComplex(v any) (c complex128, ok bool) {
	if p, ok := v.(complex64); ok {
		return complex128(p), ok
	}
	c, ok = v.(complex128)
	return
}

func IsNumeric(v any) bool {
	_, isInt := ToInt(v)
	_, isFloat := ToFloat(v)
	_, isComplex := ToComplex(v)
	return isInt || isFloat || isComplex
}

type IMul interface {
	Mul(a any) (any, error)
}

type IRMul interface {
	RMul(a any) (any, error)
}

type IAdd interface {
	Add(a any) (any, error)
}

type IRAdd interface {
	RAdd(a any) (any, error)
}

type ISub interface {
	Sub(a any) (any, error)
}

type IRSub interface {
	RSub(a any) (any, error)
}

type IDiv interface {
	Div(a any) (any, error)
}

type IRDiv interface {
	RDiv(a any) (any, error)
}

type IFloorDiv interface {
	FloorDiv(a any) (any, error)
}

type IRFloorDiv interface {
	RFloorDiv(a any) (any, error)
}

type IPow interface {
	Pow(a any) (any, error)
}

type IRPow interface {
	RPow(a any) (any, error)
}

type IEq interface {
	Eq(a any) (any, error)
}

type INe interface {
	Ne(a any) (any, error)
}

type ILt interface {
	Lt(a any) (any, error)
}

type ILe interface {
	Le(a any) (any, error)
}

type IGt interface {
	Gt(a any) (any, error)
}

type IGe interface {
	Ge(a any) (any, error)
}

type IBool interface {
	Bool() (bool, error)
}

type IPos interface {
	Pos() (any, error)
}

type INeg interface {
	Neg() (any, error)
}

type IContains interface {
	Contains(a any) (bool, error)
}

func Mul(a any, b any) (any, error) {
	if imul, ok := a.(IMul); ok {
		return imul.Mul(b)
	}
	if irmul, ok := b.(IRMul); ok {
		return irmul.RMul(b)
	}
	if IsNumeric(a) {
		if IsNumeric(b) {
			return multiplyNumeric(a, b), nil
		}
		if aI, ok := ToInt(a); ok {
			switch reflect.TypeOf(b).Kind() {
			case reflect.Slice, reflect.Array:
				return mulSliceByInt(b, aI), nil
			case reflect.String:
				return mulStrByInt(b.(string), aI), nil
			}
		}
	} else {
		if bI, ok := ToInt(b); ok {
			switch reflect.TypeOf(a).Kind() {
			case reflect.Slice, reflect.Array:
				return mulSliceByInt(a, bI), nil
			case reflect.String:
				return mulStrByInt(a.(string), bI), nil
			}
		}
	}
	return nil, fmt.Errorf("given elements are not multipliable")
}

func Add(a any, b any) (any, error) {
	if iadd, ok := a.(IAdd); ok {
		return iadd.Add(b)
	}
	if irAdd, ok := b.(IRAdd); ok {
		return irAdd.RAdd(b)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return addNumeric(a, b), nil

	}
	if bothString(a, b) {
		return a.(string) + b.(string), nil
	}
	if slices.Contains([]reflect.Kind{reflect.Slice, reflect.Array}, reflect.TypeOf(a).Kind()) &&
		slices.Contains([]reflect.Kind{reflect.Slice, reflect.Array}, reflect.TypeOf(b).Kind()) {
		return addSlices(a, b), nil
	}

	return nil, fmt.Errorf("given elements are not additible")
}

func Sub(a any, b any) (any, error) {
	if i, ok := a.(ISub); ok {
		return i.Sub(b)
	}
	if ir, ok := b.(IRSub); ok {
		return ir.RSub(b)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return subNumeric(a, b), nil
	}

	return nil, fmt.Errorf("given elements are not subable")
}

func Div(a any, b any) (any, error) {
	if i, ok := a.(IDiv); ok {
		return i.Div(b)
	}
	if ir, ok := b.(IRDiv); ok {
		return ir.RDiv(b)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return divNumeric(a, b)
	}

	return nil, fmt.Errorf("given elements are not divable")
}

func Pow(a any, b any) (any, error) {
	if i, ok := a.(IPow); ok {
		return i.Pow(b)
	}
	if ir, ok := b.(IRPow); ok {
		return ir.RPow(b)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return powNumeric(a, b), nil
	}

	return nil, fmt.Errorf("given elements are not powable")
}

func FloorDiv(a any, b any) (any, error) {
	if i, ok := a.(IFloorDiv); ok {
		return i.FloorDiv(b)
	}
	if ir, ok := b.(IRFloorDiv); ok {
		return ir.RFloorDiv(b)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return floorDivNumeric(a, b)
	}

	return nil, fmt.Errorf("given elements are not floor divable")
}

func Eq(a any, b any) (any, error) {
	if i, ok := a.(IEq); ok {
		return i.Eq(b)
	}
	if i, ok := b.(IEq); ok {
		return i.Eq(a)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return eqNumeric(a, b), nil
	}

	return reflect.DeepEqual(a, b), nil
}

func Ne(a any, b any) (any, error) {
	if i, ok := a.(INe); ok {
		return i.Ne(b)
	}
	if i, ok := b.(INe); ok {
		return i.Ne(a)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return neNumeric(a, b), nil
	}

	return !reflect.DeepEqual(a, b), nil
}

func Ge(a any, b any) (any, error) {
	if i, ok := a.(IGe); ok {
		return i.Ge(b)
	}
	if i, ok := b.(ILe); ok {
		return i.Le(a)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return geNumeric(a, b), nil
	}
	if bothString(a, b) {
		return a.(string) > b.(string), nil
	}

	return nil, fmt.Errorf("given elements are not geable")
}

func Le(a any, b any) (any, error) {
	if i, ok := a.(ILe); ok {
		return i.Le(b)
	}
	if i, ok := b.(IGe); ok {
		return i.Ge(a)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return leNumeric(a, b), nil
	}
	if bothString(a, b) {
		return a.(string) < b.(string), nil
	}

	return nil, fmt.Errorf("given elements are not floor leable")
}

func Lt(a any, b any) (any, error) {
	if i, ok := a.(ILt); ok {
		return i.Lt(b)
	}
	if i, ok := b.(IGt); ok {
		return i.Gt(a)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return ltNumeric(a, b), nil
	}
	if bothString(a, b) {
		return a.(string) <= b.(string), nil
	}

	return nil, fmt.Errorf("given elements are not floor ltable")
}

func Gt(a any, b any) (any, error) {
	if i, ok := a.(IGt); ok {
		return i.Gt(b)
	}
	if i, ok := b.(ILt); ok {
		return i.Lt(a)
	}
	if IsNumeric(a) && IsNumeric(b) {
		return gtNumeric(a, b), nil
	}
	if bothString(a, b) {
		return a.(string) >= b.(string), nil
	}

	return nil, fmt.Errorf("given elements are not floor gtable")
}

func Bool(a any) (bool, error) {
	if i, ok := a.(IBool); ok {
		return i.Bool()
	}
	return reflect.ValueOf(a).IsZero(), nil
}

func Not(a any) (bool, error) {
	b, err := Bool(a)
	if err != nil {
		return b, err
	}
	return !b, nil
}

func Pos(a any) (any, error) {
	if i, ok := a.(IPos); ok {
		return i.Pos()
	}
	if IsNumeric(a) {
		return a, nil
	}
	return nil, fmt.Errorf("given element is not posable")
}

func Neg(a any) (any, error) {
	if i, ok := a.(INeg); ok {
		return i.Neg()
	}
	if IsNumeric(a) {
		return multiplyNumeric(a, -1), nil
	}
	return nil, fmt.Errorf("given element is not negable")
}

func Contains(a, b any) (bool, error) {
	if i, ok := a.(IContains); ok {
		return i.Contains(b)
	}
	if bothString(a, b) {
		return strings.Contains(a.(string), b.(string)), nil
	}
	switch reflect.TypeOf(a).Kind() {
	case reflect.Array, reflect.Slice:
		aV := reflect.ValueOf(a)
		for i := 0; i < aV.Len(); i++ {
			r, err := Eq(aV.Index(i).Interface(), b)
			if err != nil {
				return false, err
			}
			if r == true {
				return true, nil
			}
		}
		return false, nil
	case reflect.Map:
		aV := reflect.ValueOf(a)
		iter := aV.MapRange()
		for iter.Next() {
			r, err := Eq(iter.Key().Interface(), b)
			if err != nil {
				return false, err
			}
			if r == true {
				return true, nil
			}
		}
		return false, nil
	}
	return false, fmt.Errorf("elements are not cointainable ")
}

func bothString(a, b any) bool {
	_, aOk := a.(string)
	_, bOk := b.(string)
	return aOk && bOk
}

func addSlices(a, b any) []interface{} {
	aV := reflect.ValueOf(a)
	bV := reflect.ValueOf(b)

	ret := make([]interface{}, 0, aV.Len()+bV.Len())
	for i := 0; i < aV.Len(); i++ {
		ret = append(ret, aV.Index(i).Interface())
	}
	for i := 0; i < bV.Len(); i++ {
		ret = append(ret, bV.Index(i).Interface())
	}
	return ret
}

func mulStrByInt(a string, b int64) string {
	res := ""
	for i := int64(0); i < b; i++ {
		res += a
	}
	return res
}

func mulSliceByInt(a any, b int64) interface{} {
	t, v := reflect.TypeOf(a), reflect.ValueOf(a)
	c := reflect.MakeSlice(t, v.Len()*int(b), v.Len()*int(b))
	for i := 0; i < int(b); i++ {
		reflect.Copy(c.Slice(i*v.Len(), (i+1)*v.Len()), v)
	}
	return c.Interface()
}

func multiplyNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v * b.(int64), nil
		case float64:
			return v * b.(float64), nil
		case complex128:
			return v * b.(complex128), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func addNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v + b.(int64), nil
		case float64:
			return v + b.(float64), nil
		case complex128:
			return v + b.(complex128), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func subNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v - b.(int64), nil
		case float64:
			return v - b.(float64), nil
		case complex128:
			return v - b.(complex128), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func powNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return int64(math.Pow(float64(v), float64(b.(int64)))), nil
		case float64:
			return math.Pow(v, b.(float64)), nil
		case complex128:
			// TODO impement
			return nil, fmt.Errorf("pow on imaginary numbers is not implemented")
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func divNumeric(a any, b any) (any, error) {
	return opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			bI := b.(int64)
			if bI == 0 {
				return nil, fmt.Errorf("div by 0")
			}
			return v / bI, nil
		case float64:
			bF := b.(float64)
			if bF == 0 {
				return nil, fmt.Errorf("div by 0")
			}
			return v / bF, nil
		case complex128:
			bC := b.(complex128)
			if bC == 0 {
				return nil, fmt.Errorf("div by 0")
			}
			return v / bC, nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
}

func floorDivNumeric(a any, b any) (any, error) {
	return opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			bI := b.(int64)
			if bI == 0 {
				return nil, fmt.Errorf("div by 0")
			}
			return v / bI, nil
		case float64:
			bF := b.(float64)
			if bF == 0 {
				return nil, fmt.Errorf("div by 0")
			}
			return math.Floor(v / bF), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
}

func modNumeric(a any, b any) (any, error) {
	return opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			bI := b.(int64)
			if bI == 0 {
				return nil, fmt.Errorf("modulo by 0")
			}
			return v % bI, nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
}

func eqNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v == b.(int64), nil
		case float64:
			return v == b.(float64), nil
		case complex128:
			return v == b.(complex128), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func neNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v != b.(int64), nil
		case float64:
			return v != b.(float64), nil
		case complex128:
			return v != b.(complex128), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func leNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v < b.(int64), nil
		case float64:
			return v < b.(float64), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func ltNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v <= b.(int64), nil
		case float64:
			return v <= b.(float64), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func geNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v > b.(int64), nil
		case float64:
			return v > b.(float64), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func gtNumeric(a any, b any) any {
	res, _ := opNumeric(a, b, func(a any, b any) (any, error) {
		switch v := a.(type) {
		case int64:
			return v >= b.(int64), nil
		case float64:
			return v >= b.(float64), nil
		default:
			return nil, fmt.Errorf("wrong type")
		}
	})
	return res
}

func opNumeric(a any, b any, op func(a any, b any) (any, error)) (any, error) {
	if i, ok := ToInt(a); ok {
		if i2, ok2 := ToInt(b); ok2 {
			return op(i, i2)
		}
		if f2, ok2 := ToFloat(b); ok2 {
			return op(float64(i), f2)
		}
		if c2, ok2 := ToComplex(b); ok2 {
			return op(complex(float64(i), 0), c2)
		}
	}
	if f, ok := ToFloat(a); ok {
		if i2, ok2 := ToInt(b); ok2 {
			return op(f, float64(i2))
		}
		if f2, ok2 := ToFloat(b); ok2 {
			return op(f, f2)
		}
		if c2, ok2 := ToComplex(b); ok2 {
			return op(complex(f, 0), c2)
		}
	}
	if c, ok := ToComplex(a); ok {
		if i2, ok2 := ToInt(b); ok2 {
			return op(c, complex(float64(i2), 0))
		}
		if f2, ok2 := ToFloat(b); ok2 {
			return op(c, complex(f2, 0))
		}
		if c2, ok2 := ToComplex(b); ok2 {
			return op(c, c2)
		}
	}
	return nil, fmt.Errorf("wrong type")
}
