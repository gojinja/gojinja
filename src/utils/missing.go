package utils

type missing struct{}

type Missing interface {
	Missing()
}

func (missing) Missing() {}

func GetMissing() Missing {
	return missing{}
}

func IsMissing(v any) bool {
	_, ok := v.(Missing)
	return ok
}
