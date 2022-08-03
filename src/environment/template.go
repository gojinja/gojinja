package environment

import (
	"fmt"
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/utils"
	"github.com/gojinja/gojinja/src/utils/iterator"
)

type ITemplate interface {
	IsUpToDate() bool
	Globals() map[string]any
	Render(variables map[string]any) (any, error)
	Srender(variables map[string]any) (string, error)                     // Srender, like fmt.Sprint
	Generate(variables map[string]any) (iterator.Iterator[string], error) // TODO or iterator.Iterator[any]?
	Stream(variables map[string]any) (*TemplateStream, error)
}

type Template struct {
	filename       *string
	ast            *nodes.Template
	env            *Environment
	globals        map[string]any
	upToDate       UpToDate
	isNative       bool
	name           *string // TODO track where this should come from
	blocks         map[string]stringGenerator
	rootRenderFunc map[string]stringGenerator
}

type UpToDate = func() bool

func FromString(env *Environment, source string, filename *string, globals map[string]any, upToDate UpToDate) (ITemplate, error) {
	ast, err := env.parse(source, filename)
	if err != nil {
		return nil, err
	}

	return &Template{
		filename: filename,
		ast:      ast,
		env:      env,
		globals:  env.MakeGlobals(globals),
		upToDate: upToDate,
		isNative: env.IsNative, // TODO respect this setting during rendering etc.
	}, nil
}

func (t *Template) IsUpToDate() bool {
	return t.upToDate() // TODO probably need to pass some data/context to this function?
}

func (t *Template) Globals() map[string]any {
	return t.globals
}

func (t *Template) Render(variables map[string]any) (any, error) {
	ctx := t.newContext(variables, false, nil)

	templateGenerator, err := renderTemplate(ctx, t.ast)
	if err != nil {
		return nil, err
	}
	pieces, err := iterator.ToSlice(iterator.Map(templateGenerator, utils.AsAny[string]))
	if err != nil {
		return nil, err
	}
	return t.env.Concat(pieces), nil
}

func (t *Template) Srender(variables map[string]any) (string, error) {
	s, err := t.Render(variables)
	if err != nil {
		return "", err
	}

	switch v := s.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprint(v), nil
	}
}

func (t *Template) Generate(variables map[string]any) (iterator.Iterator[string], error) {
	ctx := t.newContext(variables, false, nil)
	return renderTemplate(ctx, t.ast)
}

func (t *Template) Stream(variables map[string]any) (*TemplateStream, error) {
	it, err := t.Generate(variables)
	if err != nil {
		return nil, err
	}
	return newTemplateStream(it), nil
}

func (t *Template) newContext(variables map[string]any, shared bool, locals map[string]any) *renderContext {
	//        """Create a new :class:`Context` for this template.  The vars
	//        provided will be passed to the template.  Per default the globals
	//        are added to the context.  If shared is set to `True` the data
	//        is passed as is to the context without adding the globals.
	//
	//        `locals` can be a dict of local variables for internal usage.
	//        """
	return NewContext(t.env, t.name, t.blocks, variables, shared, t.globals, locals)
}
