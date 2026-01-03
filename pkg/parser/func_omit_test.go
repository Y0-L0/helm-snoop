package parser

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
)

func (s *Unittest) TestParseFile_Omit() {
	cases := []struct {
		name     string
		template string
		expected path.Paths
	}{
		{
			name:     "omit_simple",
			template: `{{ omit .Values.config "enabled" }}`,
			expected: path.Paths{
				path.NewPath("config"),
			},
		},
		{
			name:     "omit_piped_to_toYaml",
			template: `{{ omit .Values.containerSecurityContext "enabled" | toYaml }}`,
			expected: path.Paths{
				np().Key("containerSecurityContext").Wildcard(),
			},
		},
		{
			name:     "omit_multiple_keys",
			template: `{{ omit .Values.config "enabled" "debug" "verbose" }}`,
			expected: path.Paths{
				path.NewPath("config"),
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
