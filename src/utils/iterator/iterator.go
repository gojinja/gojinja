package iterator

// An Iterator is a stream of items of some type.
type Iterator[T any] interface {
	// Next fetches the next item in the stream.
	Next() T

	// HasNext returns true if the iterator has more items to fetch.
	HasNext() bool
}

type sliceIterator[T any] struct {
	slice []T
}

func (iter *sliceIterator[T]) HasNext() bool {
	return len(iter.slice) > 0
}

func (iter *sliceIterator[T]) Next() T {
	if len(iter.slice) == 0 {
		var v T
		return v // Anything better?
	}
	item := iter.slice[0]
	iter.slice = iter.slice[1:]
	return item
}

// FromSlice creates a new iterator which returns all items from the slice starting at index 0 until
// all items are consumed.
func FromSlice[T any](slice []T) Iterator[T] {
	return &sliceIterator[T]{slice: slice}
}

// ToSlice collects the items from the specified iterator into a slice.
func ToSlice[T any](from Iterator[T]) []T {
	var slice []T
	for from.HasNext() {
		slice = append(slice, from.Next())
	}
	return slice
}
