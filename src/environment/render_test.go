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
		input: "{{ 2 + 2 * 2 }}",
		res:   "6",
	},
	{
		input: `{% if name != "OFF" %}
	my name is {{ name }}
	{% endif %}
	{{ 5 + 1 }}`,
		res: "my name is gojinja\n6",
		variables: map[string]any{
			"name": "gojinja",
		},
	},
}

func TestRenderE2E(t *testing.T) {
	env, err := New(DefaultEnvOpts())
	if err != nil {
		t.Fatalf("Failed to create environment %v", err)
	}

	for _, c := range cases {
		tmpl, err := FromString(env, c.input, nil, c.globals, func() bool { return true })
		if err != nil {
			t.Fatalf("Failed to create template %v", err)
		}

		out, err := tmpl.Render(c.variables)
		if err != nil {
			if c.shouldErr {
				continue
			}
			t.Fatalf("Failed to render template %v", err)
		}
		if err == nil && c.shouldErr {
			t.Fatalf("Expected error during render")
		}

		if out != c.res {
			t.Fatalf("Expected %q got %q (template %q)", c.res, out, c.input)
		}
	}
}
