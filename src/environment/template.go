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
	filename *string
	ast      *nodes.Template
	env      *Environment
	globals  map[string]any
	upToDate UpToDate
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

func (t Template) IsUpToDate() bool {
	return t.upToDate() // TODO probably need to pass some data/context to this function?
}

func (t Template) Globals() map[string]any {
	return t.globals
}

func (t Template) Render(variables map[string]any) (any, error) {
	//TODO implement me
	panic("implement me")
}
