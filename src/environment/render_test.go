package environment

import (
	"testing"
)

type testRender struct {
	input     string
	res       any
	globals   map[string]any
	variables map[string]any
	shouldErr bool
}

var cases = []testRender{
	{
		input: "2137",
		res:   "2137",
	},
	{
		input: "{{ 2137 }}",
		res:   "2137",
	},
	{
		input: "1 + 1",
		res:   "1 + 1",
	},
	{
		input: "{{ 1 + 1 }}",
		res:   "2",
	},
	//	{
	//		input: `{% if name != "OFF" %}
	//my name is {{ name }}
	//{% endif %}
	//{{ 5 + 1 }}`,
	//		res: "my name is gojinja\n6",
	//		variables: map[string]any{
	//			"name": "gojinja",
	//		},
	//	},
}

func TestRenderE2E(t *testing.T) {
	env, err := New(DefaultEnvOpts())
	if err != nil {
		t.Errorf("Failed to create environment %v", err)
	}

	for _, c := range cases {
		tmpl, err := env.TemplateClass.FromString(env, c.input, nil, c.globals, func() bool { return true })
		if err != nil {
			t.Errorf("Failed to create template %v", err)
		}

		out, err := tmpl.Render(c.variables)
		if err != nil {
			if c.shouldErr {
				continue
			}
			t.Errorf("Failed to render template %v", err)
		}
		if err == nil && c.shouldErr {
			t.Error("Expected error during render")
		}

		if out != c.res {
			t.Errorf("Expected %q got %q", c.res, out)
		}
	}
}
