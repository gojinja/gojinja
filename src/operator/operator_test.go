package operator

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"reflect"
	"testing"
)

type binCase struct {
	a   any
	b   any
	res any
	err bool
}

type iAdd struct{}

func (i iAdd) Add(a any) (any, error) {
	if !IsNumeric(a) {
		return nil, fmt.Errorf("expected numeric")
	}
	return addNumeric(21, a), nil
}

var _ IAdd = iAdd{}

type iRAdd struct{}

func (i iRAdd) RAdd(a any) (any, error) {
	if !IsNumeric(a) {
		return nil, fmt.Errorf("expected numeric")
	}
	return addNumeric(21, a), nil
}

var _ IRAdd = iRAdd{}

func TestAdd(t *testing.T) {
	runBinTestCases(t, Add, []binCase{
		{3, 4, int64(7), false},
		{3, 4., float64(7), false},
		{3, complex(0, 1), complex(3, 1), false},
		{"foo", "bar", "foobar", false},
		{[]string{"foo"}, []int{3}, []any{"foo", 3}, false},
		{iAdd{}, 21, int64(42), false},
		{3, iRAdd{}, int64(24), false},
		{"3", iRAdd{}, nil, true},
		{iAdd{}, "21", nil, true},
		{"foo", 3, nil, true},
		{iRAdd{}, 21, nil, true},
		{3, iAdd{}, nil, true},
		{[]string{}, 3, nil, true},
	})
}

type iMul struct{}

func (i iMul) Mul(a any) (any, error) {
	if !IsNumeric(a) {
		return nil, fmt.Errorf("expected numeric")
	}
	return multiplyNumeric(2, a), nil
}

var _ IMul = iMul{}

type iRMul struct{}

func (i iRMul) RMul(a any) (any, error) {
	if !IsNumeric(a) {
		return nil, fmt.Errorf("expected numeric")
	}
	return multiplyNumeric(2, a), nil
}

var _ IRMul = iRMul{}

func TestMul(t *testing.T) {
	runBinTestCases(t, Mul, []binCase{
		{3, 4, int64(12), false},
		{3, 4., float64(12), false},
		{3, complex(3, 1), complex(9, 3), false},
		{[]any{"foo", 42}, 3, []any{"foo", 42, "foo", 42, "foo", 42}, false},
		{"aa", 3, "aaaaaa", false},
		{3, []any{"foo", 42}, []any{"foo", 42, "foo", 42, "foo", 42}, false},
		{3, "aa", "aaaaaa", false},
		{iMul{}, 21, int64(42), false},
		{3, iRMul{}, int64(6), false},
		{"3", iRMul{}, nil, true},
		{iMul{}, "21", nil, true},
		{iRMul{}, 21, nil, true},
		{3, iMul{}, nil, true},
		{[]string{"foo"}, []int{3}, nil, true},
		{"foo", "bar", nil, true},
	})
}

type iEq struct{}

func (i iEq) Eq(a any) (any, error) {
	return a == 42, nil
}

var _ IEq = iEq{}

func TestEq(t *testing.T) {
	runBinTestCases(t, Eq, []binCase{
		{3, 3, true, false},
		{3, 3., true, false},
		{3, complex(3, 0), true, false},
		{true, true, true, false},
		{42, iEq{}, true, false},
		{iEq{}, 42, true, false},
		{"foo", "foo", true, false},
		{[]string{"foo"}, []string{"foo"}, true, false},
		{"foo", "bar", false, false},
		{[]string{"foo"}, []string{"bar"}, false, false},
	})
}

type iLe struct{}

func (i iLe) Le(a any) (any, error) {
	if !IsNumeric(a) {
		return nil, fmt.Errorf("expected numeric")
	}
	return leNumeric(42, a), nil
}

var _ ILe = iLe{}

type iGe struct{}

func (i iGe) Ge(a any) (any, error) {
	if !IsNumeric(a) {
		return nil, fmt.Errorf("expected numeric")
	}
	return geNumeric(42, a), nil
}

var _ IGe = iGe{}

func TestLe(t *testing.T) {
	runBinTestCases(t, Le, []binCase{
		{3, 4, true, false},
		{3, 4., true, false},
		{iLe{}, 43, true, false},
		{21, iGe{}, true, false},
		{21, iLe{}, false, true},
		{iGe{}, 21, false, true},
		{"foo", "foobar", true, false},
		{true, true, false, true},
		{3, "4", false, true},
	})
}

func runBinTestCases(t *testing.T, binOp func(any, any) (any, error), cases []binCase) {
	for _, c := range cases {
		res, err := binOp(c.a, c.b)
		if err != nil {
			if c.err {
				continue
			}
			t.Fatal(err, spew.Sprint(c))
		} else if c.err {
			t.Fatal("expected error, got:", res, spew.Sprint(c))
		}
		if !reflect.DeepEqual(res, c.res) {
			t.Fatal("got:", spew.Sprint(res), ",expected:", spew.Sprint(c.res), spew.Sprint(c))
		}
	}
}
