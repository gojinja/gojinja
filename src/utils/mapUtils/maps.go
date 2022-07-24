package mapUtils

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"sort"
)

func SortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := maps.Keys(m)
	sort.Slice(keys, func(i int, j int) bool { return keys[i] < keys[j] })
	return keys
}

func Chain[K comparable, V any](ms ...map[K]V) map[K]V {
	res := make(map[K]V)
	for _, m := range ms {
		for k, v := range m {
			if _, ok := res[k]; !ok {
				res[k] = v
			}
		}
	}
	return res
}
