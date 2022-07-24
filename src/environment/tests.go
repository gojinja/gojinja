package environment

import (
	"fmt"
	"github.com/gojinja/gojinja/src/runtime"
	"github.com/gojinja/gojinja/src/utils/numbers"
	"github.com/gojinja/gojinja/src/utils/slices"
	"reflect"
	"strings"
)

func getInt(idx int, values ...any) (int64, error) {
	if len(values) <= idx {
		return 0, fmt.Errorf("not enough values passed to the test")
	}
	res, ok := numbers.ToInt(values[idx])
	if !ok {
		return 0, fmt.Errorf("value passed to the test is not an integer")
	}
	return res, nil
}

func toString(v any) string {
	return fmt.Sprint(v)
}

func testOdd(_ *Environment, f any, _ ...any) (bool, error) {
	value, ok := numbers.ToInt(f)
	if !ok {
		return false, fmt.Errorf("value passed to the test is not an integer")
	}
	return value%2 == 1, nil
}

func testEven(_ *Environment, f any, _ ...any) (bool, error) {
	value, ok := numbers.ToInt(f)
	if !ok {
		return false, fmt.Errorf("value passed to the test is not an integer")
	}
	return value%2 == 0, nil
}

func testDivisibleBy(_ *Environment, f any, values ...any) (bool, error) {
	value, ok := numbers.ToInt(f)
	if !ok {
		return false, fmt.Errorf("value passed to the test is not an integer")
	}
	num, err := getInt(0, values...)
	if err != nil {
		return false, err
	}
	if num == 0 {
		return false, fmt.Errorf("tried to divide by zero")
	}
	return value%num == 0, nil
}

func testUndefined(_ *Environment, value any, _ ...any) (bool, error) {
	_, ok := value.(runtime.IUndefined)
	return ok, nil
}

func testDefined(e *Environment, value any, values ...any) (bool, error) {
	un, err := testUndefined(e, value, values...)
	return !un, err
}

func testFilter(env *Environment, value any, _ ...any) (bool, error) {
	v, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("argument is not a string")
	}
	_, in := env.Filters[v]
	return in, nil
}

func testTest(env *Environment, value any, _ ...any) (bool, error) {
	v, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("argument is not a string")
	}
	_, in := env.Tests[v]
	return in, nil
}

func testNone(_ *Environment, value any, _ ...any) (bool, error) {
	return value == nil, nil
}

func testBoolean(_ *Environment, value any, _ ...any) (bool, error) {
	_, isBool := value.(bool)
	return isBool, nil
}

func testFalse(_ *Environment, value any, _ ...any) (bool, error) {
	return value == false, nil
}

func testTrue(_ *Environment, value any, _ ...any) (bool, error) {
	return value == true, nil
}

func testInteger(_ *Environment, value any, _ ...any) (bool, error) {
	_, ok := numbers.ToInt(value)
	return ok, nil
}

func testFloat(_ *Environment, value any, _ ...any) (bool, error) {
	_, ok := numbers.ToFloat(value)
	return ok, nil
}

func testLower(_ *Environment, value any, _ ...any) (bool, error) {
	s := toString(value)
	return strings.ToLower(s) == s, nil
}

func testUpper(_ *Environment, value any, _ ...any) (bool, error) {
	s := toString(value)
	return strings.ToUpper(s) == s, nil
}

func testString(_ *Environment, value any, _ ...any) (bool, error) {
	_, ok := value.(string)
	return ok, nil
}

func testMapping(_ *Environment, value any, _ ...any) (bool, error) {
	return reflect.TypeOf(value).Kind() == reflect.Map, nil
}

func testNumber(_ *Environment, value any, _ ...any) (bool, error) {
	return numbers.IsNumeric(value), nil
}

func testSequence(_ *Environment, value any, _ ...any) (bool, error) {
	// TODO rewrite using operator len and getitem
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return true, nil
	default:
		return false, nil
	}
}

func testCallable(_ *Environment, value any, _ ...any) (bool, error) {
	// TODO rewrite using operator call
	return reflect.TypeOf(value).Kind() == reflect.Func, nil
}

func testSameAs(_ *Environment, value any, values ...any) (bool, error) {
	// It is not exactly the same as in jinja as it's impossible to be as jinja uses `is`.
	if len(values) == 0 {
		return false, fmt.Errorf("not enough values passed to the function")
	}
	v2 := values[0]
	switch reflect.TypeOf(value).Kind() {
	case reflect.Bool:
		return value == v2, nil
	case reflect.Chan, reflect.Map, reflect.Func, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		if slices.Contains([]reflect.Kind{reflect.Chan, reflect.Map, reflect.Func, reflect.Pointer, reflect.Slice, reflect.UnsafePointer}, reflect.TypeOf(v2).Kind()) {
			return reflect.ValueOf(value).Pointer() == reflect.ValueOf(v2).Pointer(), nil
		}
		return false, nil
	default:
		return false, nil
	}
}

func testIterable(_ *Environment, value any, _ ...any) (bool, error) {
	// TODO rewrite using operator iter
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		return true, nil
	default:
		return false, nil
	}
}

type Escaped interface {
	HTML() (string, error)
}

func testEscaped(_ *Environment, value any, _ ...any) (bool, error) {
	_, ok := value.(Escaped)
	return ok, nil
}

func testIn(_ *Environment, value any, values ...any) (bool, error) {
	// TODO rewrite using operator contains
	if len(values) == 0 {
		return false, fmt.Errorf("not enough values passed to the function")
	}

	switch reflect.TypeOf(values[0]).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(values[0])
		for i := 0; i < s.Len(); i++ {
			if value == s.Index(i).Interface() {
				return true, nil
			}
		}
		return false, nil
	case reflect.Map:
		s := reflect.ValueOf(values[0])
		v := s.MapIndex(reflect.ValueOf(value))
		return v != reflect.Value{}, nil
	case reflect.String:
		v := values[0].(string)
		if vS, ok := value.(string); ok {
			return strings.Contains(v, vS), nil
		}
		return false, nil
	default:
		return false, fmt.Errorf("second arguemnt is not a slice")
	}
}

// Test represents a test function. Some tests only require one variable
type Test func(env *Environment, firstArg any, args ...any) (bool, error)

var Default = map[string]Test{
	"odd":         testOdd,
	"even":        testEven,
	"divisibleby": testDivisibleBy,
	"defined":     testDefined,
	"undefined":   testUndefined,
	"filter":      testFilter,
	"test":        testTest,
	"none":        testNone,
	"boolean":     testBoolean,
	"false":       testFalse,
	"true":        testTrue,
	"integer":     testInteger,
	"float":       testFloat,
	"lower":       testLower,
	"upper":       testUpper,
	"string":      testString,
	"mapping":     testMapping,
	"number":      testNumber,
	"sequence":    testSequence,
	"iterable":    testIterable,
	"in":          testIn,
	"callable":    testCallable,
	"sameas":      testSameAs,
	"escaped":     testEscaped,
	// TODO operators
}
