package parser

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
)

// Helper for building paths in tests
func np() *path.Path {
	return &path.Path{}
}

// TestParseFile_Range tests range statement evaluation
func (s *Unittest) TestParseFile_Range() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "range_over_values_path",
			template: `{{ range .Values.items }}{{ end }}`,
			expected: path.Paths{},
		},
		{
			name:     "range_body_accesses_values",
			template: `{{ range .Values.items }}{{ .Values.name }}{{ end }}`,
			expected: path.Paths{path.NewPath("name")},
		},
		{
			name:     "range_with_else",
			template: `{{ range .Values.items }}{{ else }}{{ .Values.fallback }}{{ end }}`,
			expected: path.Paths{path.NewPath("fallback")},
		},
		{
			name: "range_body_and_else_both_access_values",
			template: `{{ range .Values.items }}{{ .Values.itemName }}` +
				`{{ else }}{{ .Values.defaultName }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("itemName"),
				path.NewPath("defaultName"),
			},
		},
		{
			name: "nested_range",
			template: `{{ range .Values.outer }}{{ range .Values.inner }}` +
				`{{ .Values.leaf }}{{ end }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("leaf"),
			},
		},
		{
			name:     "range_with_variable_assignment",
			template: `{{ range $item := .Values.items }}{{ .Values.name }}{{ end }}`,
			expected: path.Paths{path.NewPath("name")},
		},
		{
			name:     "range_with_index_and_value",
			template: `{{ range $index, $item := .Values.items }}{{ .Values.name }}{{ end }}`,
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

// TestParseFile_RangePrefix tests that range blocks change the context
// so that .foo inside "range .Values.items" refers to items[*].foo
func (s *Unittest) TestParseFile_RangePrefix() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "range_changes_dot_context",
			template: `{{ range .Values.items }}{{ .name }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
			},
		},
		{
			name:     "range_multiple_field_access",
			template: `{{ range .Values.users }}{{ .id }}{{ .email }}{{ end }}`,
			expected: path.Paths{
				np().Key("users").Wildcard().Key("id"),
				np().Key("users").Wildcard().Key("email"),
			},
		},
		{
			name:     "range_deep_field_access",
			template: `{{ range .Values.services }}{{ .metadata.name }}{{ end }}`,
			expected: path.Paths{
				np().Key("services").Wildcard().Key("metadata").Key("name"),
			},
		},
		{
			name: "nested_range_contexts",
			template: `{{ range .Values.teams }}
				{{ .name }}
				{{ range .members }}
					{{ .email }}
				{{ end }}
			{{ end }}`,
			expected: path.Paths{
				np().Key("teams").Wildcard().Key("name"),
				np().Key("teams").Wildcard().Key("members").Wildcard().Key("email"),
			},
		},
		{
			name: "range_else_preserves_original_context",
			template: `{{ range .Values.items }}{{ .name }}` +
				`{{ else }}{{ .Values.emptyMessage }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
				path.NewPath("emptyMessage"),
			},
		},
		{
			name:     "range_dollar_accesses_root_context",
			template: `{{ range .Values.items }}{{ $.Values.config }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config"),
			},
		},
		// TODO: Variable tracking not yet implemented
		// {
		// 	name:     "range_with_variable_still_tracks_field_access",
		// 	template: `{{ range $item := .Values.items }}{{ $item.name }}{{ end }}`,
		// 	expected: path.Paths{
		// 		path.NewPath("items"),
		// 		path.NewPath("items", "*", "name"),
		// 	},
		// },
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			actual, err := parseFile(tc.name+".tmpl", []byte(tc.template), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.expected, actual)
		})
	}
}

// TestParseFile_RangeWithInteraction tests range and with together with other features
func (s *Unittest) TestParseFile_RangeWithInteraction() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name: "range_inside_if",
			template: `{{ if .Values.enabled }}{{ range .Values.items }}` +
				`{{ .Values.name }}{{ end }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("enabled"),
				path.NewPath("name"),
			},
		},
		{
			name: "with_inside_range",
			template: `{{ range .Values.items }}{{ with .Values.config }}` +
				`{{ .Values.setting }}{{ end }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("setting"),
			},
		},
		{
			name:     "range_with_function_call",
			template: `{{ range .Values.items }}{{ .Values.name | upper }}{{ end }}`,
			expected: path.Paths{path.NewPath("name")},
		},
		{
			name: "with_with_function_call",
			template: `{{ with .Values.config | default .Values.fallback }}` +
				`{{ .Values.name }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config"),
				path.NewPath("fallback"),
				path.NewPath("name"),
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			if tc.name == "with_with_function_call" {
				s.T().Skip("Skipped: defaultFn returns paths instead of emitting")
			}
			actual, err := parseFile(tc.name+".tmpl", []byte(tc.template), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.expected, actual)
		})
	}
}

// TestParseFile_RangeWithMixedContexts tests interaction between range/with and explicit .Values
func (s *Unittest) TestParseFile_RangeWithMixedContexts() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "with_explicit_values_overrides_context",
			template: `{{ with .Values.config }}{{ .name }}{{ .Values.global }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "name"),
				path.NewPath("global"),
			},
		},
		{
			name:     "range_explicit_values_overrides_context",
			template: `{{ range .Values.items }}{{ .name }}{{ .Values.count }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
				path.NewPath("count"),
			},
		},
		{
			name: "with_inside_range",
			template: `{{ range .Values.items }}
				{{ with .config }}
					{{ .timeout }}
				{{ end }}
			{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("config").Key("timeout"),
			},
		},
		{
			name: "range_inside_with",
			template: `{{ with .Values.app }}
				{{ range .services }}
					{{ .port }}
				{{ end }}
			{{ end }}`,
			expected: path.Paths{
				np().Key("app").Key("services").Wildcard().Key("port"),
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

// TestParseFile_RangeBuiltinObjects tests that built-in Helm objects like
// Chart and Release are never tracked as .Values paths, even when accessed
// from within a range block that sets a prefix.
func (s *Unittest) TestParseFile_RangeBuiltinObjects() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "chart_not_tracked_in_range",
			template: `{{ range .Values.items }}{{ .Chart.AppVersion }}{{ end }}`,
			expected: path.Paths{},
		},
		{
			name:     "release_not_tracked_in_range",
			template: `{{ range .Values.items }}{{ .Release.Name }}{{ end }}`,
			expected: path.Paths{},
		},
		{
			name:     "chart_and_release_not_tracked",
			template: `{{ range .Values.items }}{{ .Chart.AppVersion }}{{ .Release.Name }}{{ end }}`,
			expected: path.Paths{},
		},
		{
			name:     "builtin_with_relative_paths",
			template: `{{ range .Values.items }}{{ .name }}{{ .Chart.AppVersion }}{{ end }}`,
			expected: path.Paths{
				np().Key("items").Wildcard().Key("name"),
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
