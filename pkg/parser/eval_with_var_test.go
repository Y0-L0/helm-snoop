package parser

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
)

// TestParseFile_WithVariables tests with variable tracking
// Phase 3: With block variables
func (s *Unittest) TestParseFile_WithVariables() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		// Basic with variable
		{
			name:     "with_variable_field_access",
			template: `{{ with $cfg := .Values.config }}{{ $cfg.enabled }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "enabled"),
			},
		},
		{
			name:     "with_variable_nested_field_access",
			template: `{{ with $cfg := .Values.config }}{{ $cfg.db.host }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "db", "host"),
			},
		},
		{
			name:     "with_variable_multiple_fields",
			template: `{{ with $cfg := .Values.config }}{{ $cfg.port }}{{ $cfg.host }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "port"),
				path.NewPath("config", "host"),
			},
		},

		// With variable and direct access
		{
			name:     "with_variable_and_direct_values",
			template: `{{ with $cfg := .Values.config }}{{ $cfg.port }}{{ .Values.global }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "port"),
				path.NewPath("global"),
			},
		},

		// No variable - existing behavior should still work
		{
			name:     "with_no_variable_uses_dot_context",
			template: `{{ with .Values.config }}{{ .enabled }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "enabled"),
			},
		},

		// Bare variable
		{
			name:     "with_bare_variable_reference",
			template: `{{ with $cfg := .Values.config }}{{ $cfg }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config"),
			},
		},

		// Variable not in else
		{
			name: "with_variable_not_in_else",
			template: `{{ with $cfg := .Values.config }}{{ $cfg.enabled }}` +
				`{{ else }}{{ .Values.fallback }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "enabled"),
				path.NewPath("fallback"),
			},
		},

		// Nested with blocks
		{
			name: "nested_with_different_vars",
			template: `{{ with $outer := .Values.config }}{{ $outer.port }}` +
				`{{ with $inner := .Values.db }}{{ $inner.host }}{{ end }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "port"),
				path.NewPath("db", "host"),
			},
		},
		{
			name: "nested_with_outer_var_accessible",
			template: `{{ with $outer := .Values.config }}` +
				`{{ with $inner := .Values.db }}{{ $outer.port }}{{ $inner.host }}{{ end }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "port"),
				path.NewPath("db", "host"),
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
