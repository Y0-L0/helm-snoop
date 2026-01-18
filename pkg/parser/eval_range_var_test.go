package parser

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
)

// TestParseFile_RangeVariables tests range variable tracking
// Phase 1: Basic range variable usage
func (s *Unittest) TestParseFile_RangeVariables() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		// Single variable assignment
		{
			name:     "single_variable_field_access",
			template: `{{ range $item := .Values.items }}{{ $item.name }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
			},
		},
		{
			name:     "single_variable_nested_field_access",
			template: `{{ range $item := .Values.items }}{{ $item.config.enabled }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("config").Key("enabled"),
			},
		},
		{
			name:     "single_variable_multiple_fields",
			template: `{{ range $item := .Values.items }}{{ $item.name }}{{ $item.value }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
				np().Key("items").Wildcard().Key("value"),
			},
		},

		// Key-value variable assignment
		{
			name:     "key_value_variables_value_access",
			template: `{{ range $key, $value := .Values.ports }}{{ $value.port }}{{ end }}`,
			expected: path.Paths{
				np().Key("ports").Wildcard().Key("port"),
			},
		},
		{
			name:     "key_value_variables_multiple_fields",
			template: `{{ range $k, $v := .Values.ports }}{{ $v.port }}{{ $v.protocol }}{{ end }}`,
			expected: path.Paths{
				np().Key("ports").Wildcard().Key("port"),
				np().Key("ports").Wildcard().Key("protocol"),
			},
		},
		{
			name:     "key_value_variables_nested_access",
			template: `{{ range $k, $v := .Values.service.ports }}{{ $v.config.enabled }}{{ end }}`,
			expected: path.Paths{
				np().Key("service").Key("ports").Wildcard().Key("config").Key("enabled"),
			},
		},

		// Key access should not be tracked (keys are strings, not paths)
		{
			name:     "key_variable_not_tracked",
			template: `{{ range $key, $value := .Values.ports }}{{ $key }}{{ end }}`,
			expected: path.Paths{},
		},

		// Mixed: variable and direct .Values access
		{
			name:     "variable_and_direct_values",
			template: `{{ range $item := .Values.items }}{{ $item.name }}{{ .Values.global }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
				path.NewPath("global"),
			},
		},
		{
			name:     "variable_and_root_context",
			template: `{{ range $item := .Values.items }}{{ $item.name }}{{ $.Values.root }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
				path.NewPath("root"),
			},
		},

		// Bare variable (just $item without field access)
		{
			name:     "bare_variable_reference",
			template: `{{ range $item := .Values.items }}{{ $item }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard(),
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			actual, err := parseFile("", tc.name+".tmpl", []byte(tc.template), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.expected, actual)
		})
	}
}

// TestParseFile_RangeVariablesNested tests nested range with variables
// Phase 2: Complex scenarios
func (s *Unittest) TestParseFile_RangeVariablesNested() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		// Nested ranges with different variable names
		{
			name: "nested_ranges_different_vars",
			template: `{{ range $outer := .Values.items }}{{ $outer.name }}` +
				`{{ range $inner := .Values.protocols }}{{ $inner.type }}{{ end }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
				np().Key("protocols").Wildcard().Key("type"),
			},
		},
		{
			name: "nested_ranges_outer_var_accessible",
			template: `{{ range $outer := .Values.items }}` +
				`{{ range $inner := .Values.protocols }}{{ $outer.name }}{{ $inner.type }}{{ end }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
				np().Key("protocols").Wildcard().Key("type"),
			},
		},

		// Variable shadowing
		{
			name: "nested_ranges_same_var_name_shadowing",
			template: `{{ range $item := .Values.outer }}{{ $item.name }}` +
				`{{ range $item := .Values.inner }}{{ $item.name }}{{ end }}{{ end }}`,
			expected: path.Paths{
				np().Key("outer").Wildcard().Key("name"),
				np().Key("inner").Wildcard().Key("name"),
			},
		},
		{
			name: "shadowing_inner_overrides_outer",
			template: `{{ range $item := .Values.outer }}` +
				`{{ range $item := .Values.inner }}{{ $item.value }}{{ end }}` +
				`{{ $item.name }}{{ end }}`,
			expected: path.Paths{
				np().Key("inner").Wildcard().Key("value"),
				np().Key("outer").Wildcard().Key("name"),
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			actual, err := parseFile("", tc.name+".tmpl", []byte(tc.template), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.expected, actual)
		})
	}
}

// TestParseFile_RangeVariablesMixed tests range variables with other features
// Phase 2 continued: Mixed usage
func (s *Unittest) TestParseFile_RangeVariablesMixed() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		// Variables with if statements
		{
			name:     "variable_in_if_condition",
			template: `{{ range $item := .Values.items }}{{ if $item.enabled }}yes{{ end }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("enabled"),
			},
		},
		{
			name: "variable_in_if_body",
			template: `{{ range $item := .Values.items }}` +
				`{{ if .Values.global }}{{ $item.name }}{{ end }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("global"),
				np().Key("items").Wildcard().Key("name"),
			},
		},

		// Variables are not available in else block
		{
			name: "range_variable_not_in_else",
			template: `{{ range $item := .Values.items }}{{ $item.name }}` +
				`{{ else }}{{ .Values.fallback }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
				path.NewPath("fallback"),
			},
		},

		// No variable - existing behavior should still work
		{
			name:     "no_variable_uses_dot_context",
			template: `{{ range .Values.items }}{{ .name }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
			},
		},
		{
			name: "no_variable_and_with_variable_mixed",
			template: `{{ range .Values.outer }}{{ .name }}{{ end }}` +
				`{{ range $item := .Values.inner }}{{ $item.value }}{{ end }}`,
			expected: path.Paths{
				np().Key("outer").Wildcard().Key("name"),
				np().Key("inner").Wildcard().Key("value"),
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			actual, err := parseFile("", tc.name+".tmpl", []byte(tc.template), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.expected, actual)
		})
	}
}
