package operator

import (
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type callCase struct {
	fun  any
	args []any
	ret  []any
	err  bool
}

type iCall struct{}

func (iCall) Call(args ...any) ([]any, error) {
	return args, nil
}

var _ ICall = iCall{}

func TestCall(t *testing.T) {
	zeroArg := func() (any, error) {
		return 42, nil
	}
	oneArg := func(h int) int {
		return h
	}
	oneArgAny := func(h any) any {
		return h
	}

	variadic := func(p int, rest ...string) string {
		return strconv.Itoa(p) + "|" + strings.Join(rest, "|")
	}

	cases := []callCase{
		{0, nil, nil, true},
		{zeroArg, nil, []any{42, nil}, false},
		{zeroArg, []any{0}, nil, true},
		{oneArg, nil, nil, true},
		{oneArg, []any{42}, []any{42}, false},
		{oneArgAny, []any{42}, []any{42}, false},
		{oneArgAny, []any{"foo"}, []any{"foo"}, false},
		{oneArg, []any{"foo"}, nil, true},
		{oneArg, []any{6, 9}, nil, true},
		{variadic, nil, nil, true},
		{variadic, []any{42}, []any{"42|"}, false},
		{variadic, []any{"42"}, nil, true},
		{variadic, []any{42, "6", "9"}, []any{"42|6|9"}, false},
		{variadic, []any{42, "6", 9}, nil, true},
		{iCall{}, []any{42, "6", 9}, []any{42, "6", 9}, false},
	}
	for _, c := range cases {
		handleCallCase(t, c)
	}
}

func handleCallCase(t *testing.T, c callCase) {
	res, err := Call(c.fun, c.args...)
	if err != nil {
		if c.err {
			return
		}
		t.Fatal(err, spew.Sprint(c))
	} else if c.err {
		t.Fatal("expected error, got:", res, spew.Sprint(c))
	}
	if !reflect.DeepEqual(res, c.ret) {
		t.Fatal("got:", spew.Sprint(res), ",expected:", spew.Sprint(c.ret), spew.Sprint(c))
	}
}
