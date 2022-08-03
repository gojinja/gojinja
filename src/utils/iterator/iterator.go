package iterator

// An Iterator is a stream of items of some type.
type Iterator[T any] interface {
	// Next fetches the next item in the stream.
	Next() (T, error)

	// HasNext returns true if the iterator has more items to fetch.
	HasNext() bool
}

type ExhaustedError struct{}

func (e ExhaustedError) Error() string {
	return "iterator has no more items"
}

type sliceIterator[T any] struct {
	slice []T
}

func (iter *sliceIterator[T]) HasNext() bool {
	return len(iter.slice) > 0
}

func (iter *sliceIterator[T]) Next() (T, error) {
	if len(iter.slice) == 0 {
		var zero T
		return zero, ExhaustedError{}
	}
	item := iter.slice[0]
	iter.slice = iter.slice[1:]
	return item, nil
}

// FromSlice creates a new iterator which returns all items from the slice starting at index 0 until
// all items are consumed.
func FromSlice[T any](slice []T) Iterator[T] {
	return &sliceIterator[T]{slice: slice}
}

// Once creates a new iterator which returns the single item passed as an argument.
func Once[T any](item T) Iterator[T] {
	return &sliceIterator[T]{slice: []T{item}}
}

// ToSlice collects the items from the specified iterator into a slice.
func ToSlice[T any](from Iterator[T]) ([]T, error) {
	var slice []T
	for from.HasNext() {
		r, err := from.Next()
		if err != nil {
			return nil, err
		}
		slice = append(slice, r)
	}
	return slice, nil
}

func Map[T, F any](it Iterator[T], f func(T) (F, error)) Iterator[F] {
	return &mapIterator[T, F]{
		it: it,
		f: func(r T) (F, error) {
			v, err := f(r)
			if err != nil {
				var zero F
				return zero, err
			}
			return v, nil
		},
	}
}

type mapIterator[T, F any] struct {
	it Iterator[T]
	f  func(T) (F, error)
}

func (iter *mapIterator[T, F]) HasNext() bool {
	return iter.it.HasNext()
}

func (iter *mapIterator[T, F]) Next() (F, error) {
	v, err := iter.it.Next()
	if err != nil {
		var zero F
		return zero, err
	}
	return iter.f(v)
}
