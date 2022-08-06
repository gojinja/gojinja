package nodes

import "github.com/gojinja/gojinja/src/utils/iterator"

func Find[T Node](node Node) *T {
	it := FindAll[T](node)
	if !it.HasNext() {
		return nil
	}
	found, err := it.Next()
	if err != nil {
		// This should never happen - error in walking over AST is a bug in gojinja.
		panic(err)
	}
	return &found
}

func FindAll[T Node](node Node) iterator.Iterator[T] {
	found := iterator.Empty[T]()

	children := node.IterChildNodes(nil, nil)
	for children.HasNext() {
		child, _ := children.Next() // Error in walking over AST would be a bug in gojinja.
		if v, ok := child.(T); ok {
			found = iterator.Chain(found, iterator.Once(v))
		}
		found = iterator.Chain(found, FindAll[T](child))
	}

	return found
}
