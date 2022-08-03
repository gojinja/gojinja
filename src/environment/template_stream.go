package environment

import (
	"fmt"
	"github.com/gojinja/gojinja/src/utils/iterator"
	"strings"
)

type TemplateStream struct {
	// TODO im 99% sure this should generate strings, not interface{}, but if I'm wrong it will be trivial to fix in the future.
	internalIterator iterator.Iterator[string]
	bufferSize       int
}

func newTemplateStream(it iterator.Iterator[string]) *TemplateStream {
	return &TemplateStream{
		internalIterator: it,
		bufferSize:       1,
	}
}

func (t *TemplateStream) Next() (string, error) {
	if t.bufferSize == 1 {
		return t.internalIterator.Next()
	}

	var buffer []string
	nonemptyItems := 0
	for t.internalIterator.HasNext() && nonemptyItems < t.bufferSize {
		c, err := t.internalIterator.Next()
		if err != nil {
			return "", err
		}
		buffer = append(buffer, c)
		if c != "" {
			nonemptyItems += 1
		}
	}

	// This is literal translation from jinja -- and most likely a bug in jinja.
	// default concat("") is used instead of environment.Concat
	return strings.Join(buffer, ""), nil
}

func (t *TemplateStream) HasNext() bool {
	return t.internalIterator.HasNext()
}

func (t *TemplateStream) DisableBuffering() {
	t.bufferSize = 1
}

func (t *TemplateStream) EnableBuffering(size int) error {
	if size <= 1 {
		return fmt.Errorf("buffer size must be greater than 1")
	}
	t.bufferSize = size
	return nil
}

func (t *TemplateStream) Dump(file, encoding, errors any) error {
	// TODO
	panic("not implemented")
}

// Make sure TemplateStream satisfies the iterator.Iterator interface.
var _ iterator.Iterator[string] = &TemplateStream{}
