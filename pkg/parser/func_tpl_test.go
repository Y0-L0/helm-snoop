package parser

import (
	"github.com/y0-l0/helm-snoop/internal/assert"
	"github.com/y0-l0/helm-snoop/pkg/path"
)

// TestParseFile_TplFunction tests basic tpl function support
func (s *Unittest) TestParseFile_TplFunction() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "tpl_with_simple_values_path",
			template: `{{ tpl .Values.postgresql.auth.username . }}`,
			expected: path.Paths{
				path.NewPath("postgresql", "auth", "username"),
			},
		},
		{
			name: "tpl_in_range_context",
			template: `{{ range .Values.imagePullSecrets }}` +
				`{{ tpl . $ }}{{ end }}`,
			expected: path.Paths{
				np().Key("imagePullSecrets").Wildcard(),
			},
		},
		{
			name: "tpl_in_with_context",
			template: `{{ with .Values.config }}` +
				`{{ tpl .template . }}{{ end }}`,
			expected: path.Paths{
				path.NewPath("config", "template"),
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			// Disable strict mode for tpl tests since it's only partially implemented
			oldStrict := assert.Strict
			assert.Strict = false
			defer func() { assert.Strict = oldStrict }()

			actual, err := parseFile("", tc.name+".tmpl", []byte(tc.template), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.expected, actual)
		})
	}
}
