package tplparser

import (
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

func (s *Unittest) TestParseFile_Omit() {
	cases := []struct {
		name     string
		template string
		expected vpath.Paths
	}{
		{
			name:     "omit_simple",
			template: `{{ omit .Values.config "enabled" }}`,
			expected: vpath.Paths{
				vpath.NewPath("config"),
			},
		},
		{
			name:     "omit_piped_to_toYaml",
			template: `{{ omit .Values.containerSecurityContext "enabled" | toYaml }}`,
			expected: vpath.Paths{
				np().Key("containerSecurityContext").Wildcard(),
			},
		},
		{
			name:     "omit_multiple_keys",
			template: `{{ omit .Values.config "enabled" "debug" "verbose" }}`,
			expected: vpath.Paths{
				vpath.NewPath("config"),
			},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			vpath.EqualPaths(s, tc.expected, s.parse(tc.template))
		})
	}
}
