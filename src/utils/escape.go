package utils

import "fmt"

func Escape(v any) string {
	// TODO port markupsafe.escape
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", v)
	}
}
