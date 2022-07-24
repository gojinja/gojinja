package numbers

import "fmt"

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
	if isInt {
		return true
	}
	_, isFloat := ToFloat(v)
	if isFloat {
		return true
	}
	_, isComplex := ToComplex(v)
	return isComplex
}
