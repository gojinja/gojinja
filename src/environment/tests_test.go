package environment

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/gojinja/gojinja/src/filters"
	"github.com/gojinja/gojinja/src/runtime"
	"testing"
)

type testCase struct {
	env  *Environment
	f    any
	rest []any
	res  bool
	err  bool
}

func TestOdd(t *testing.T) {
	runTestCases(t, testOdd, []testCase{
		{nil, 21, nil, true, false},
		{nil, 1337, nil, true, false},
		{nil, 69, nil, true, false},
		{nil, 22, nil, false, false},
		{nil, 1338, nil, false, false},
		{nil, 70, nil, false, false},
		{nil, "", nil, false, true},
		{nil, &struct{}{}, nil, false, true},
	})
}

func TestEven(t *testing.T) {
	runTestCases(t, testEven, []testCase{
		{nil, 21, nil, false, false},
		{nil, 1337, nil, false, false},
		{nil, 69, nil, false, false},
		{nil, 22, nil, true, false},
		{nil, 1338, nil, true, false},
		{nil, 70, nil, true, false},
		{nil, "", nil, false, true},
		{nil, &struct{}{}, nil, false, true},
	})
}

func TestDivisibleBy(t *testing.T) {
	runTestCases(t, testDivisibleBy, []testCase{
		{nil, 42, []any{2}, true, false},
		{nil, 6, []any{9}, false, false},
		{nil, 2, []any{0}, false, true},
		{nil, 2, nil, false, true},
		{nil, "", []any{1}, false, true},
		{nil, "", []any{""}, false, true},
	})
}

func TestDefined(t *testing.T) {
	runTestCases(t, testDefined, []testCase{
		{nil, 0, nil, true, false},
		{nil, "", nil, true, false},
		{nil, runtime.NewUndefined(nil, nil, nil, nil, nil), nil, false, false},
		{nil, runtime.NewChainableUndefined(nil, nil, nil, nil, nil), nil, false, false},
		{nil, runtime.NewStrictUndefined(nil, nil, nil, nil, nil), nil, false, false},
		{nil, runtime.NewDebugUndefined(nil, nil, nil, nil, nil), nil, false, false},
	})
}

func TestUndefined(t *testing.T) {
	runTestCases(t, testUndefined, []testCase{
		{nil, 0, nil, false, false},
		{nil, "", nil, false, false},
		{nil, runtime.NewUndefined(nil, nil, nil, nil, nil), nil, true, false},
		{nil, runtime.NewChainableUndefined(nil, nil, nil, nil, nil), nil, true, false},
		{nil, runtime.NewStrictUndefined(nil, nil, nil, nil, nil), nil, true, false},
		{nil, runtime.NewDebugUndefined(nil, nil, nil, nil, nil), nil, true, false},
	})
}

func testEnv() *Environment {
	env, _ := New(DefaultEnvOpts())
	env.Filters = map[string]filters.Filter{
		"bar": nil,
	}
	env.Tests = map[string]Test{
		"foo": nil,
	}
	return env
}

func TestFilter(t *testing.T) {
	runTestCases(t, testFilter, []testCase{
		{testEnv(), "foo", nil, false, false},
		{testEnv(), "bar", nil, true, false},
	})
}

func TestTest(t *testing.T) {
	runTestCases(t, testTest, []testCase{
		{testEnv(), "foo", nil, true, false},
		{testEnv(), "bar", nil, false, false},
	})
}

func TestNone(t *testing.T) {
	runTestCases(t, testNone, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, false, false},
		{nil, 0, nil, false, false},
		{nil, nil, nil, true, false},
	})
}

func TestBoolean(t *testing.T) {
	runTestCases(t, testBoolean, []testCase{
		{nil, true, nil, true, false},
		{nil, false, nil, true, false},
		{nil, 0, nil, false, false},
	})
}

func TestFalse(t *testing.T) {
	runTestCases(t, testFalse, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, true, false},
		{nil, 0, nil, false, false},
	})
}

func TestTrue(t *testing.T) {
	runTestCases(t, testTrue, []testCase{
		{nil, true, nil, true, false},
		{nil, false, nil, false, false},
		{nil, 1, nil, false, false},
	})
}

func TestInteger(t *testing.T) {
	runTestCases(t, testInteger, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, false, false},
		{nil, 1, nil, true, false},
		{nil, int8(1), nil, true, false},
		{nil, int16(1), nil, true, false},
		{nil, int32(1), nil, true, false},
		{nil, int64(1), nil, true, false},
		{nil, uint8(1), nil, true, false},
		{nil, uint16(1), nil, true, false},
		{nil, uint32(1), nil, true, false},
		{nil, uint64(1), nil, true, false},
		{nil, uintptr(1), nil, false, false},
		{nil, complex(0, 0), nil, false, false},
		{nil, 0., nil, false, false},
		{nil, nil, nil, false, false},
		{nil, &struct{}{}, nil, false, false},
	})
}

func TestFloat(t *testing.T) {
	runTestCases(t, testFloat, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, false, false},
		{nil, 1, nil, false, false},
		{nil, int8(1), nil, false, false},
		{nil, int16(1), nil, false, false},
		{nil, int32(1), nil, false, false},
		{nil, int64(1), nil, false, false},
		{nil, uint8(1), nil, false, false},
		{nil, uint16(1), nil, false, false},
		{nil, uint32(1), nil, false, false},
		{nil, uint64(1), nil, false, false},
		{nil, uintptr(1), nil, false, false},
		{nil, complex(0, 0), nil, false, false},
		{nil, float32(0.), nil, true, false},
		{nil, 0., nil, true, false},
		{nil, &struct{}{}, nil, false, false},
	})
}

func TestLower(t *testing.T) {
	runTestCases(t, testLower, []testCase{
		{nil, 1, nil, true, false},
		{nil, "foo bar", nil, true, false},
		{nil, "fOo bAR", nil, false, false},
		{nil, "FOO BAR", nil, false, false},
		{nil, "", nil, true, false},
	})
}

func TestUpper(t *testing.T) {
	runTestCases(t, testUpper, []testCase{
		{nil, 1, nil, true, false},
		{nil, "foo bar", nil, false, false},
		{nil, "fOo bAR", nil, false, false},
		{nil, "FOO BAR", nil, true, false},
		{nil, "", nil, true, false},
	})
}

func TestString(t *testing.T) {
	runTestCases(t, testString, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, false, false},
		{nil, 1, nil, false, false},
		{nil, int8(1), nil, false, false},
		{nil, int16(1), nil, false, false},
		{nil, int32(1), nil, false, false},
		{nil, int64(1), nil, false, false},
		{nil, uint8(1), nil, false, false},
		{nil, uint16(1), nil, false, false},
		{nil, uint32(1), nil, false, false},
		{nil, uint64(1), nil, false, false},
		{nil, uintptr(1), nil, false, false},
		{nil, complex(0, 0), nil, false, false},
		{nil, float32(0.), nil, false, false},
		{nil, 0., nil, false, false},
		{nil, &struct{}{}, nil, false, false},
		{nil, 'a', nil, false, false},
		{nil, []byte("foo"), nil, false, false},
		{nil, "foo", nil, true, false},
		{nil, "", nil, true, false},
	})
}

func TestMapping(t *testing.T) {
	runTestCases(t, testMapping, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, false, false},
		{nil, 1, nil, false, false},
		{nil, int8(1), nil, false, false},
		{nil, int16(1), nil, false, false},
		{nil, int32(1), nil, false, false},
		{nil, int64(1), nil, false, false},
		{nil, uint8(1), nil, false, false},
		{nil, uint16(1), nil, false, false},
		{nil, uint32(1), nil, false, false},
		{nil, uint64(1), nil, false, false},
		{nil, uintptr(1), nil, false, false},
		{nil, complex(0, 0), nil, false, false},
		{nil, float32(0.), nil, false, false},
		{nil, 0., nil, false, false},
		{nil, &struct{}{}, nil, false, false},
		{nil, 'a', nil, false, false},
		{nil, []byte("foo"), nil, false, false},
		{nil, "foo", nil, false, false},
		{nil, "", nil, false, false},
		{nil, map[string]string{}, nil, true, false},
		{nil, map[any]any{}, nil, true, false},
		{nil, map[string]string{"foo": "bar"}, nil, true, false},
	})
}

func TestNumber(t *testing.T) {
	runTestCases(t, testNumber, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, false, false},
		{nil, 1, nil, true, false},
		{nil, int8(1), nil, true, false},
		{nil, int16(1), nil, true, false},
		{nil, int32(1), nil, true, false},
		{nil, int64(1), nil, true, false},
		{nil, uint8(1), nil, true, false},
		{nil, uint16(1), nil, true, false},
		{nil, uint32(1), nil, true, false},
		{nil, uint64(1), nil, true, false},
		{nil, uintptr(1), nil, false, false},
		{nil, complex(0, 0), nil, true, false},
		{nil, complex64(complex(0, 0)), nil, true, false},
		{nil, float32(0.), nil, true, false},
		{nil, 0., nil, true, false},
		{nil, &struct{}{}, nil, false, false},
		{nil, 'a', nil, true, false},
		{nil, []byte("foo"), nil, false, false},
		{nil, "foo", nil, false, false},
		{nil, "", nil, false, false},
	})
}

func TestSequence(t *testing.T) {
	runTestCases(t, testSequence, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, false, false},
		{nil, 1, nil, false, false},
		{nil, int8(1), nil, false, false},
		{nil, int16(1), nil, false, false},
		{nil, int32(1), nil, false, false},
		{nil, int64(1), nil, false, false},
		{nil, uint8(1), nil, false, false},
		{nil, uint16(1), nil, false, false},
		{nil, uint32(1), nil, false, false},
		{nil, uint64(1), nil, false, false},
		{nil, uintptr(1), nil, false, false},
		{nil, complex(0, 0), nil, false, false},
		{nil, complex64(complex(0, 0)), nil, false, false},
		{nil, float32(0.), nil, false, false},
		{nil, 0., nil, false, false},
		{nil, &struct{}{}, nil, false, false},
		{nil, 'a', nil, false, false},
		{nil, "foo", nil, true, false},
		{nil, "", nil, true, false},
		{nil, []byte("foo"), nil, true, false},
		{nil, []byte("foo")[2:], nil, true, false},
		{nil, map[string]string{}, nil, true, false},
		{nil, []any{6, 9, 42}, nil, true, false},
		{nil, make(chan string), nil, false, false},
	})
}

func TestIterable(t *testing.T) {
	runTestCases(t, testIterable, []testCase{
		{nil, true, nil, false, false},
		{nil, false, nil, false, false},
		{nil, 1, nil, false, false},
		{nil, int8(1), nil, false, false},
		{nil, int16(1), nil, false, false},
		{nil, int32(1), nil, false, false},
		{nil, int64(1), nil, false, false},
		{nil, uint8(1), nil, false, false},
		{nil, uint16(1), nil, false, false},
		{nil, uint32(1), nil, false, false},
		{nil, uint64(1), nil, false, false},
		{nil, uintptr(1), nil, false, false},
		{nil, complex(0, 0), nil, false, false},
		{nil, complex64(complex(0, 0)), nil, false, false},
		{nil, float32(0.), nil, false, false},
		{nil, 0., nil, false, false},
		{nil, &struct{}{}, nil, false, false},
		{nil, 'a', nil, false, false},
		{nil, "foo", nil, true, false},
		{nil, "", nil, true, false},
		{nil, []byte("foo"), nil, true, false},
		{nil, []byte("foo")[2:], nil, true, false},
		{nil, map[string]string{}, nil, true, false},
		{nil, []any{6, 9, 42}, nil, true, false},
		{nil, make(chan string), nil, true, false},
	})
}

func TestIn(t *testing.T) {
	runTestCases(t, testIn, []testCase{
		{nil, "foo", []any{[]string{"foo", "bar"}}, true, false},
		{nil, "bar", []any{[]string{"foo", "bar"}}, true, false},
		{nil, "", []any{[]string{"foo", "bar"}}, false, false},
		{nil, "a", []any{"Ala"}, true, false},
		{nil, 'a', []any{"Ala"}, false, true},
		{nil, "o", []any{"Ala"}, false, false},
		{nil, 'o', []any{0}, false, true},
		{nil, "o", []any{map[string]string{"o": "o"}}, true, false},
		{nil, "a", []any{map[string]string{"o": "o"}}, false, false},
	})
}

func TestCallable(t *testing.T) {
	runTestCases(t, testCallable, []testCase{
		{nil, "foo", nil, false, false},
		{nil, 0, nil, false, false},
		{nil, struct{}{}, nil, false, false},
		{nil, func() {}, nil, true, false},
		{nil, func(...any) any { return 0 }, nil, true, false},
	})
}

func TestSameAs(t *testing.T) {
	a := 0
	b := 0

	runTestCases(t, testSameAs, []testCase{
		{nil, true, []any{true}, true, false},
		{nil, true, []any{false}, false, false},
		{nil, false, []any{false}, true, false},
		{nil, false, []any{true}, false, false},
		{nil, false, []any{0}, false, false},
		{nil, struct{}{}, []any{struct{}{}}, false, false},
		{nil, &a, []any{&b}, false, false},
		{nil, &a, []any{&a}, true, false},
		{nil, &b, []any{&b}, true, false},
	})
}

type escaped struct{}

func (escaped) HTML() (string, error) { return "HTML", nil }

var _ Escaped = escaped{}

func TestEscaped(t *testing.T) {
	runTestCases(t, testEscaped, []testCase{
		{nil, "foo", nil, false, false},
		{nil, 0, nil, false, false},
		{nil, struct{}{}, nil, false, false},
		{nil, func() {}, nil, false, false},
		{nil, func(...any) any { return 0 }, nil, false, false},
		{nil, escaped{}, nil, true, false},
		{nil, struct{}{}, nil, false, false},
	})
}

func runTestCases(t *testing.T, test Test, cases []testCase) {
	for _, c := range cases {
		res, err := test(c.env, c.f, c.rest...)
		if err != nil {
			if !c.err {
				t.Fatal(err)
			} else {
				continue
			}
		} else if c.err {
			t.Fatal("expected error", spew.Sprint(c))
		}
		if res != c.res {
			t.Fatal("got: ", res, ", expected: ", c.res, spew.Sprint(c))
		}
	}
}
