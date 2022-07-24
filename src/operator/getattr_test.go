package operator

import (
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
)

type cleanStruct struct {
	Foo string
}

type getAttrStruct struct {
	Foo string
}

func (g getAttrStruct) GetAttr(name string) (any, error) {
	return name, nil
}

var _ IGetAttr = getAttrStruct{}

type getAttributeStruct struct {
	Foo string
}

func (g getAttributeStruct) GetAttribute(name string) (any, error) {
	return name, nil
}

var _ IGetAttribute = getAttributeStruct{}

type getAttrCase struct {
	s    any
	name string
	res  any
	err  bool
}

func TestGetAttr(t *testing.T) {
	cases := []getAttrCase{
		{cleanStruct{"bar"}, "Foo", "bar", false},
		{cleanStruct{"bar"}, "foo", nil, true},
		{getAttrStruct{"bar"}, "Foo", "bar", false},
		{getAttrStruct{"bar"}, "foo", "foo", false},
		{getAttributeStruct{"bar"}, "Foo", "Foo", false},
		{getAttributeStruct{"bar"}, "foo", "foo", false},
		{0, "Foo", nil, true},
	}

	for _, c := range cases {
		runGetAttrCase(t, c)
	}
}

func runGetAttrCase(t *testing.T, c getAttrCase) {
	res, err := GetAttr(c.s, c.name)
	if err != nil {
		if c.err {
			return
		}
		t.Fatal(err, spew.Sprint(c))
	} else if c.err {
		t.Fatal("expected error, got:", res, spew.Sprint(c))
	}
	if !reflect.DeepEqual(res, c.res) {
		t.Fatal("got:", spew.Sprint(res), ",expected:", spew.Sprint(c.res), spew.Sprint(c))
	}

}
