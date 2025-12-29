package parser

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
)

// TestParseFile_With tests with statement evaluation
func (s *Unittest) TestParseFile_With() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "with_values_path",
			template: `{{ with .Values.config }}{{ end }}`,
			expected: path.Paths{path.NewPath("config")},
		},
		{
			name:     "with_body_accesses_values",
			template: `{{ with .Values.config }}{{ .Values.name }}{{ end }}`,
			expected: path.Paths{path.NewPath("config"), path.NewPath("name")},
		},
		{
			name:     "with_else",
			template: `{{ with .Values.config }}{{ else }}{{ .Values.defaultConfig }}{{ end }}`,
			expected: path.Paths{path.NewPath("config"), path.NewPath("defaultConfig")},
		},
		{
			name: "with_body_and_else_both_access_values",
			template: `{{ with .Values.config }}{{ .Values.enabled }}` +
				`{{ else }}{{ .Values.disabled }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config"),
				path.NewPath("enabled"),
				path.NewPath("disabled"),
			},
		},
		{
			name: "nested_with",
			template: `{{ with .Values.outer }}{{ with .Values.inner }}` +
				`{{ .Values.leaf }}{{ end }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("outer"),
				path.NewPath("inner"),
				path.NewPath("leaf"),
			},
		},
		{
			name:     "with_variable_assignment",
			template: `{{ with $cfg := .Values.config }}{{ .Values.name }}{{ end }}`,
			expected: path.Paths{path.NewPath("config"), path.NewPath("name")},
		},
		{
			name:     "with_dot_context",
			template: `{{ with . }}{{ .Values.name }}{{ end }}`,
			expected: path.Paths{path.NewPath("name")},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			actual, err := parseFile(tc.name+".tmpl", []byte(tc.template), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.expected, actual)
		})
	}
}

// TestParseFile_WithPrefix tests that with blocks change the context
// so that .foo inside "with .Values.config" refers to .Values.config.foo
func (s *Unittest) TestParseFile_WithPrefix() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "with_changes_dot_context",
			template: `{{ with .Values.config }}{{ .name }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config"),
				path.NewPath("config", "name"),
			},
		},
		{
			name:     "with_nested_field_access",
			template: `{{ with .Values.database }}{{ .host }}{{ .port }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("database"),
				path.NewPath("database", "host"),
				path.NewPath("database", "port"),
			},
		},
		{
			name:     "with_deep_field_access",
			template: `{{ with .Values.app }}{{ .config.timeout }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("app"),
				path.NewPath("app", "config", "timeout"),
			},
		},
		{
			name: "nested_with_contexts",
			template: `{{ with .Values.outer }}
				{{ .field1 }}
				{{ with .inner }}
					{{ .field2 }}
				{{ end }}
			{{ end }}`,
			expected: path.Paths{
				path.NewPath("outer"),
				path.NewPath("outer", "field1"),
				path.NewPath("outer", "inner"),
				path.NewPath("outer", "inner", "field2"),
			},
		},
		{
			name: "with_else_preserves_original_context",
			template: `{{ with .Values.config }}{{ .name }}` +
				`{{ else }}{{ .Values.default }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config"),
				path.NewPath("config", "name"),
				path.NewPath("default"),
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			actual, err := parseFile(tc.name+".tmpl", []byte(tc.template), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.expected, actual)
		})
	}
}
