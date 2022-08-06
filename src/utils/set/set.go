package set

import "golang.org/x/exp/maps"

type Set[K comparable] map[K]struct{}

func New[K comparable]() Set[K] {
	return make(Set[K])
}

func (s Set[K]) Add(key K) {
	s[key] = struct{}{}
}

func (s Set[K]) Remove(key K) {
	delete(s, key)
}

func (s Set[K]) Has(key K) bool {
	_, ok := s[key]
	return ok
}

func FromElems[K comparable](list ...K) Set[K] {
	set := New[K]()
	for _, el := range list {
		set.Add(el)
	}
	return set
}

func (s Set[K]) Clone() Set[K] {
	return maps.Clone(s)
}
