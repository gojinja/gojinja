package utils

import "strings"

func Concat(maybeStrings []any) any {
	var strs []string
	for _, v := range maybeStrings {
		s, ok := v.(string)
		if !ok {
			panic("concatenation of non-string values using default (non native) concatenator (this is a bug in gojinja)")
		}
		strs = append(strs, s)
	}

	return strings.Join(strs, "")
}

func NativeConcat(parts []any) any {
	// TODO port from jinja (requires ported literal_eval)
	panic("not implemented")
}

func StrAsAny(v string) any {
	return v
}
