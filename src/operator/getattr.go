package operator

import (
	"fmt"
	"reflect"
)

type IGetAttribute interface {
	GetAttribute(name string) (any, error)
}

type IGetAttr interface {
	GetAttr(name string) (any, error)
}

func GetAttr(v any, name string) (any, error) {
	if gA, ok := v.(IGetAttribute); ok {
		return gA.GetAttribute(name)
	}
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Struct {
		res := value.FieldByName(name)
		if res.Kind() != reflect.Invalid {
			return res.Interface(), nil
		}
	}
	if gA, ok := v.(IGetAttr); ok {
		return gA.GetAttr(name)
	}
	return nil, fmt.Errorf("can't get attribute %s of element", name)
}
