package environment

import (
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/utils/iterator"
	"strings"
)

type output struct {
	builder strings.Builder
}

func (o *output) write(value any) error {
	switch v := value.(type) {
	case string:
		if _, err := o.builder.WriteString(v); err != nil {
			return err
		}
	default:
		panic("unsupported type")
	}
	return nil
}

type renderer struct {
	out *output
	ctx *renderContext
}

func renderTemplate(ctx *renderContext, node *nodes.Template) (iterator.Iterator[string], error) {
	// Renders the template piece by piece (iteration over statements).

	// TODO support rendering the template in pieces (iteration over statements)

	renderer := &renderer{
		ctx: ctx,
		out: &output{},
	}
	if err := renderer.renderTemplate(node); err != nil {
		return iterator.Once(""), err
	}
	return iterator.FromSlice([]string{renderer.out.builder.String()}), nil
}

func (r *renderer) renderTemplate(node *nodes.Template) error {
	if err := r.out.write("xD"); err != nil { // TODO
		return err
	}
	return nil
}
