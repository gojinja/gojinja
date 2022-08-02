package environment

import (
	"github.com/gojinja/gojinja/src/nodes"
	"strings"
)

// TODO support rendering the template in pieces (iteration over statements)

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

func renderTemplate(ctx *renderContext, node *nodes.Template) (string, error) {
	renderer := &renderer{
		ctx: ctx,
		out: &output{},
	}
	if err := renderer.renderTemplate(node); err != nil {
		return "", err
	}
	return renderer.out.builder.String(), nil
}

func (r *renderer) renderTemplate(node *nodes.Template) error {
	if err := r.out.write("xD"); err != nil { // TODO
		return err
	}
	return nil
}
