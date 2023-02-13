package operator

import (
	"fmt"
	"reflect"
)

type ICall interface {
	Call(...any) ([]any, error)
}

func Call(f any, args ...any) ([]any, error) {
	if c, ok := f.(ICall); ok {
		return c.Call(args...)
	}
	value := reflect.ValueOf(f)
	if value.Kind() != reflect.Func {
		return nil, fmt.Errorf("element is not callable")
	}

	vArgs := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		vArgs = append(vArgs, reflect.ValueOf(arg))
	}

	if value.Type().IsVariadic() {
		return handleVariadic(value, vArgs)
	}
	return handleNormal(value, vArgs)
}

func handleVariadic(value reflect.Value, args []reflect.Value) ([]any, error) {
	t := value.Type()
	numIn := t.NumIn()
	if numIn-1 > len(args) {
		return nil, fmt.Errorf("tried to call function with wrong number of arguments, got: %d, expected at least %d", len(args), numIn-1)
	}
	for i := 0; i < numIn-1; i++ {
		if !args[i].Type().AssignableTo(t.In(i)) {
			return nil, fmt.Errorf("argument number %d has a wrong type, got: '%s', expected '%s'", i, args[i].Type(), t.In(i))
		}
	}

	variadicType := t.In(numIn - 1).Elem()
	for i := numIn - 1; i < len(args); i++ {
		if !args[i].Type().AssignableTo(variadicType) {
			return nil, fmt.Errorf("variadic argument number %d has a wrong type, got: '%s', expected '%s'", i, args[i].Type(), variadicType)
		}
	}
	res := value.Call(args)
	return convertReturn(res), nil
}

func handleNormal(value reflect.Value, args []reflect.Value) ([]any, error) {
	t := value.Type()
	numIn := t.NumIn()
	if numIn != len(args) {
		return nil, fmt.Errorf("tried to call function with wrong number of arguments, got: %d, expected %d", len(args), numIn)
	}
	for i := 0; i < numIn; i++ {
		if !args[i].Type().AssignableTo(t.In(i)) {
			return nil, fmt.Errorf("argument number %d has a wrong type, got: '%s', expected '%s'", i, args[i].Type(), t.In(i))
		}
	}
	res := value.Call(args)
	return convertReturn(res), nil
}

func convertReturn(ret []reflect.Value) []any {
	r := make([]any, 0, len(ret))
	for _, v := range ret {
		r = append(r, v.Interface())
	}
	return r
}
