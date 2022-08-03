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

	renderer := &renderer{
		ctx: ctx,
		out: &output{},
	}
	if err := renderer.validateAST(node); err != nil {
		return nil, err
	}

	return iterator.Map(iterator.FromSlice(node.Body), func(n nodes.Node) (string, error) {
		if err := renderer.renderNode(n); err != nil {
			return "", err
		}
		s := renderer.out.builder.String()
		renderer.out.builder.Reset()
		return s, nil
	}), nil
}

func (r *renderer) renderNode(node nodes.Node) error {
	if err := r.out.write("xD"); err != nil { // TODO
		return err
	}
	return nil
}

func (r *renderer) validateAST(*nodes.Template) error {
	// TODO
	return nil
}
