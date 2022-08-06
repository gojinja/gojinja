package environment

import (
	"github.com/gojinja/gojinja/src/utils"
	"github.com/gojinja/gojinja/src/utils/iterator"
	"github.com/gojinja/gojinja/src/utils/set"
	"golang.org/x/exp/maps"
)

type evalContext struct {
	env        *Environment
	autoEscape bool
	volatile   bool
	store      map[string]any
	kvstore    map[string]any
}

func (e *evalContext) Get(key string) (any, bool) {
	k, ok := e.kvstore[key]
	return k, ok
}

func (e *evalContext) Set(key string, value any) {
	e.kvstore[key] = value
}

type stringGenerator = func(context *renderContext) (iterator.Iterator[string], error)

type renderContext struct {
	env          *Environment
	parent       map[string]any
	name         *string
	vars         map[string]any
	exportedVars set.Set[string]
	globalsKeys  set.Set[string]
	blocks       map[string][]stringGenerator
	evalCtx      *evalContext
}

func (c *renderContext) Get(key string, default_ any) any {
	rv := c.ResolveOrMissing(key)
	if utils.IsMissing(rv) {
		return default_
	}
	return rv
}

func (c *renderContext) Resolve(key string) any {
	// Look up a variable by name, or return an :class:`Undefined`
	// object if the key is not found.

	rv := c.ResolveOrMissing(key)
	if utils.IsMissing(rv) {
		return c.env.Undefined(nil, nil, &key, nil, nil)
	}
	return rv
}

func (c *renderContext) ResolveOrMissing(key string) any {
	// Look up a variable by name, or return a ``missing`` sentinel
	// if the key is not found.

	if v, ok := c.vars[key]; ok {
		return v
	}
	if v, ok := c.parent[key]; ok {
		return v
	}
	return utils.GetMissing()
}

func newEvalContext(env *Environment, templateName *string) *evalContext {
	return &evalContext{
		env:        env,
		autoEscape: env.AutoEscape(templateName),
		volatile:   false,
		store:      make(map[string]any),
		kvstore:    make(map[string]any),
	}
}

func newRenderContext(env *Environment, parent map[string]any, templateName *string, blocks map[string]stringGenerator, globals map[string]any) *renderContext {
	globalsKeys := set.New[string]()
	for k := range globals {
		globalsKeys.Add(k)
	}

	// create the initial mapping of blocks. Whenever template inheritance
	// takes place the runtime will update this mapping with the new blocks
	// from the template.
	wrappedBlocks := make(map[string][]stringGenerator)
	for k, v := range blocks {
		wrappedBlocks[k] = []stringGenerator{v}
	}

	return &renderContext{ // TODO use env.contextClass
		parent:       parent,
		vars:         make(map[string]any),
		env:          env,
		evalCtx:      newEvalContext(env, templateName),
		exportedVars: set.New[string](),
		name:         templateName,
		globalsKeys:  globalsKeys,
		blocks:       wrappedBlocks,
	}
}

func NewContext(env *Environment, templateName *string, blocks map[string]stringGenerator, vars map[string]any, shared bool, globals map[string]any, locals map[string]any) *renderContext {
	// TODO should probably return an interface of a context instead of concrete class renderContext
	if vars == nil {
		vars = make(map[string]any)
	}

	var parent map[string]any
	if shared {
		parent = vars
	} else {
		parent = make(map[string]any, len(globals)+len(vars))
		for k, v := range globals {
			parent[k] = v
		}
		for k, v := range vars {
			parent[k] = v
		}
	}

	if len(locals) > 0 {
		if shared {
			parent = maps.Clone(parent)
		}
		for k, v := range locals {
			if !utils.IsMissing(v) {
				parent[k] = v
			}
		}
	}

	// TODO use env.contextClass
	return newRenderContext(env, parent, templateName, blocks, globals)
}

type ContextClass struct{}
