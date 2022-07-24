package operator

import (
	"fmt"
	"reflect"
	"unicode/utf8"
)

type Iterator interface {
	Next() bool
	Elem() any
}

type IIter interface {
	Iter() (Iterator, error)
}

type arrayIter struct {
	s   []any
	idx int
}

func (i *arrayIter) Next() bool {
	i.idx += 1
	return i.idx < len(i.s)
}

func (i *arrayIter) Elem() any {
	return i.s[i.idx]
}

type MapIterEl struct {
	Key   any
	Value any
}

type mapIter struct {
	m   *reflect.MapIter
	key any
}

func (i *mapIter) Next() bool {
	if !i.m.Next() {
		return false
	}
	i.key = i.m.Key().Interface()
	return true
}

func (i *mapIter) Elem() any {
	return i.key
}

type stringIter struct {
	s string
	c rune
}

func (i *stringIter) Next() bool {
	if len(i.s) == 0 {
		return false
	}
	var size int
	i.c, size = utf8.DecodeRuneInString(i.s)
	if i.c == utf8.RuneError {
		return false
	}
	i.s = i.s[size:]
	return true
}

func (i *stringIter) Elem() any {
	return i.c
}

type chanIter struct {
	v  reflect.Value
	el any
}

func (i *chanIter) Next() bool {
	el, ok := i.v.Recv()
	if !ok {
		return false
	}
	i.el = el.Interface()
	return true
}

func (i *chanIter) Elem() any {
	return i.el
}

func Iter(a any) (Iterator, error) {
	if i, ok := a.(IIter); ok {
		return i.Iter()
	}
	value := reflect.ValueOf(a)
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		ret := make([]any, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			ret = append(ret, value.Index(i).Interface())
		}
		return &arrayIter{ret, -1}, nil
	case reflect.String:
		return &stringIter{a.(string), 0}, nil
	case reflect.Map:
		return &mapIter{value.MapRange(), nil}, nil
	case reflect.Chan:
		return &chanIter{value, nil}, nil
	default:
		return nil, fmt.Errorf("element is not iterable")
	}
}
