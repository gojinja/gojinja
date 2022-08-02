package environment

import (
	"github.com/gojinja/gojinja/src/nodes"
)

type ITemplate interface {
	IsUpToDate() bool
	Globals() map[string]any
	Render(variables map[string]any) (any, error)
}

type Class struct{}

type Template struct {
	filename       *string
	ast            *nodes.Template
	env            *Environment
	globals        map[string]any
	upToDate       UpToDate
	name           *string // TODO track where this should come from
	blocks         map[string]func(ctx *renderContext) ([]string, error)
	rootRenderFunc map[string]func(ctx *renderContext) ([]string, error)
}

type UpToDate = func() bool

func (c Class) FromString(env *Environment, source string, filename *string, globals map[string]any, upToDate UpToDate) (ITemplate, error) {
	ast, err := env.parse(source, filename)
	if err != nil {
		return nil, err
	}

	return &Template{ // TODO don't ignore `Class`
		filename: filename,
		ast:      ast,
		env:      env,
		globals:  env.MakeGlobals(globals),
		upToDate: upToDate,
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
	return renderTemplate(ctx, t.ast)
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
