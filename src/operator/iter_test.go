package operator

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
)

type iterCase struct {
	el      any
	iterRes []any
	err     bool
}

type iIter struct{}

type it struct {
	idx int
}

func (i *it) Next() bool {
	i.idx++
	return i.idx <= 3
}

func (i it) Elem() any {
	switch i.idx {
	case 1:
		return 6
	case 2:
		return 9
	case 3:
		return 42
	default:
		panic("unexpected idx")
	}
}

func (i iIter) Iter() (Iterator, error) {
	return &it{}, nil
}

var _ IIter = iIter{}

func TestIter(t *testing.T) {
	ch := make(chan any)
	go func() {
		ch <- 6
		ch <- 9
		ch <- 42
		close(ch)
	}()
	cases := []iterCase{
		{0, nil, true},
		{"foo", []any{'f', 'o', 'o'}, false},
		{[]string{"f", "o", "o"}, []any{"f", "o", "o"}, false},
		{map[string]string{"f": "1", "o": "2"}, []any{"f", "o"}, false},
		{ch, []any{6, 9, 42}, false},
		{iIter{}, []any{6, 9, 42}, false},
	}
	for _, c := range cases {
		runIterCase(t, c)
	}
}

func runIterCase(t *testing.T, c iterCase) {
	res, err := Iter(c.el)
	if err != nil {
		if c.err {
			return
		}
		t.Fatal(err, spew.Sprint(c))
	} else if c.err {
		t.Fatal("expected error", spew.Sprint(c))
	}
	i := 0
	for res.Next() {
		if i >= len(c.iterRes) {
			t.Fatal("more elements then expected", spew.Sprint(c))
		}
		if !reflect.DeepEqual(res.Elem(), c.iterRes[i]) {
			fmt.Println(reflect.TypeOf(res.Elem()), reflect.TypeOf(c.iterRes[i]))
			t.Fatal("got:", res.Elem(), ", expected:", c.iterRes[i])
		}
		i++
	}
	if i < len(c.iterRes) {
		t.Fatal("less elements then expected", spew.Sprint(c))
	}
}
