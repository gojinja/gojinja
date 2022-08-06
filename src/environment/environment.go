package environment

import (
	"fmt"
	"github.com/gojinja/gojinja/src/defaults"
	"github.com/gojinja/gojinja/src/extensions"
	"github.com/gojinja/gojinja/src/filters"
	"github.com/gojinja/gojinja/src/lexer"
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/parser"
	"github.com/gojinja/gojinja/src/runtime"
	"github.com/gojinja/gojinja/src/utils"
	"github.com/gojinja/gojinja/src/utils/mapUtils"
	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"log"
)

type ExtensionsMap map[string]extensions.IExtension

// Environment is the core component of Jinja.  It contains
// important shared variables like configuration, filters, tests,
// globals and others.  Instances of this class may be modified if
// they are not shared and if no template was loaded so far.
// Modifications on environments after the first template was loaded
// will lead to surprising effects and undefined behavior.
type Environment struct {
	Sandboxed    bool
	Overlayed    bool
	LinkedTo     *Environment
	Shared       bool
	Concat       func([]any) any
	ContextClass ContextClass
	IsNative     bool
	*lexer.EnvLexerInformation
	Optimized  bool
	Extensions ExtensionsMap
	Undefined  UndefinedConstructor
	Finalize   func(...any) any
	AutoEscape func(name *string) bool
	Loader     *Loader
	Cache      Cache
	AutoReload bool
	Filters    map[string]filters.Filter
	Tests      map[string]Test
	Globals    map[string]any
	Policies   map[string]any
}

type Cache interface {
	Add(key, value interface{}) (evicted bool)
	Get(key interface{}) (value interface{}, ok bool)
}

type mapCache map[interface{}]interface{}

func (m mapCache) Add(key, value interface{}) (evicted bool) {
	m[key] = value
	return false
}
func (m mapCache) Get(key interface{}) (value interface{}, ok bool) {
	v, ok := m[key]
	return v, ok
}

func New(opts *EnvOpts) (*Environment, error) {
	concat := utils.Concat
	if opts.IsNative {
		concat = utils.NativeConcat
	}

	env := &Environment{
		Sandboxed:           false,
		Overlayed:           false,
		LinkedTo:            nil,
		Shared:              false,
		Concat:              concat,
		ContextClass:        ContextClass{},
		EnvLexerInformation: opts.EnvLexerInformation,
		Optimized:           opts.Optimized,
		Undefined:           opts.Undefined,
		Finalize:            opts.Finalize,
		Loader:              opts.Loader,
		AutoReload:          opts.AutoReload,
		IsNative:            opts.IsNative,
		Filters:             maps.Clone(filters.Default),
		Tests:               maps.Clone(Default),
		Globals:             maps.Clone(defaults.DefaultNamespace),
		Policies:            maps.Clone(defaults.DefaultPolicies),
	}

	var err error
	env.AutoEscape, err = convertAutoEscape(opts.AutoEscape)
	if err != nil {
		return nil, err
	}

	env.Cache, err = createCache(opts.CacheSize)
	if err != nil {
		return nil, err
	}

	env.Extensions = LoadExtensions(env, opts.Extensions)
	if err = configCheck(env); err != nil {
		return nil, err
	}
	return env, nil
}

func convertAutoEscape(ae any) (func(name *string) bool, error) {
	switch v := ae.(type) {
	case bool:
		return func(*string) bool { return v }, nil
	case func(*string) bool:
		return v, nil
	default:
		return nil, fmt.Errorf("unexpected type of AutoEscape")
	}
}

func configCheck(env *Environment) error {
	if env.BlockStartString == env.VariableStartString || env.BlockStartString == env.CommentStartString || env.CommentStartString == env.VariableStartString {
		return fmt.Errorf("block, variable and comment start strings must be different")
	}
	if !slices.Contains([]string{"\n", "\r\n", "\r"}, env.NewlineSequence) {
		return fmt.Errorf("'NewlineSequence' must be one of '\\n', '\\r\\n', or '\\r'")
	}
	return nil
}

func LoadExtensions(env *Environment, extensions map[string]func(*Environment) extensions.IExtension) ExtensionsMap {
	ret := make(ExtensionsMap)
	for k, v := range extensions {
		ret[k] = v(env)
	}
	return ret
}

func createCache(cacheSize int) (Cache, error) {
	if cacheSize > 0 {
		return lru.New(cacheSize)
	}
	if cacheSize < 0 {
		return make(mapCache), nil
	}
	return nil, nil
}

type EnvOpts struct {
	*lexer.EnvLexerInformation
	Optimized  bool
	Extensions map[string]func(*Environment) extensions.IExtension // TODO jinja accepts also extensions names but it's python import magic I don't know how to do it in golang.
	Undefined  UndefinedConstructor
	Finalize   func(...any) any
	AutoEscape any // bool or func(string)bool
	Loader     *Loader
	CacheSize  int
	AutoReload bool
	IsNative   bool
}

type UndefinedConstructor func(hint *string, obj any, name *string, exc func(msg string) error, logger *log.Logger) runtime.IUndefined

func DefaultEnvOpts() *EnvOpts {
	return &EnvOpts{
		Optimized:           true,
		Extensions:          nil,
		EnvLexerInformation: lexer.DefaultEnvLexerInformation(),
		Undefined: func(hint *string, obj any, name *string, exc func(msg string) error, logger *log.Logger) runtime.IUndefined {
			return runtime.NewUndefined(hint, obj, name, exc, logger)
		},
		Finalize:   nil,
		AutoEscape: false,
		Loader:     nil,
		CacheSize:  400,
		AutoReload: true,
	}
}

// GetTemplate loads a template by name with `Loader` and returns a `Template`.
// If the template does not exist a `TemplateNotFound` exception is raised.
func (env *Environment) GetTemplate(name any, parent *string, globals map[string]any) (ITemplate, error) {
	switch v := name.(type) {
	case ITemplate:
		return v, nil
	case string:
		if parent != nil {
			v = env.JoinPath(v, *parent)
		}
		return env.loadTemplate(v, globals)
	default:
		return nil, fmt.Errorf("unexpected type for `name`")
	}
}

// JoinPath joins a template with the parent. By default, all the lookups are
// relative to the loader root so this method returns the `template`
// parameter unchanged, but if the paths should be relative to the
// parent template, this function can be used to calculate the real
// template name.
//
// Subclasses may override this method and implement template path
// joining here.
func (env *Environment) JoinPath(v string, parent string) string {
	// TODO in jinja it may be overwritten by subclass
	return v
}

func (env *Environment) loadTemplate(name string, globals map[string]any) (ITemplate, error) {
	if env.Loader == nil {
		return nil, fmt.Errorf("no loader for this environment specified")
	}
	cacheKey := fmt.Sprint(env.Loader, name)

	if env.Cache != nil {
		template, ok := env.Cache.Get(cacheKey)
		if ok {
			tmpl := template.(ITemplate)
			if !env.AutoReload && tmpl.IsUpToDate() {
				maps.Copy(tmpl.Globals(), globals)
			}
			return tmpl, nil
		}
	}
	template, err := (*env.Loader).Load(env, name, env.MakeGlobals(globals))
	if err != nil {
		return nil, err
	}
	if env.Cache != nil {
		env.Cache.Add(cacheKey, template)
	}
	return template, nil
}

func (env *Environment) MakeGlobals(globals map[string]any) map[string]any {
	return mapUtils.Chain(globals, env.Globals)
}

func (env *Environment) parse(source string, filename *string) (*nodes.Template, error) {
	// TODO stream returned from lexer should be filtered through extensions streamFilters
	stream, err := lexer.GetLexer(env.EnvLexerInformation).Tokenize(source, nil, filename, nil) // TODO fill name and state?
	if err != nil {
		return nil, err
	}
	p := parser.NewParser(stream, nil, nil, filename, nil) // TODO pass extensions slice; TODO fill name and state?
	return p.Parse()
}
