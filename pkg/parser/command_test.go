package parser

import "github.com/y0-l0/helm-snoop/pkg/path"

// quote/upper/lower simply wrap a value; default X Y reads only Y when X is literal.
func (s *Unittest) TestParseCommand_Noops() {
	cases := []struct {
		name string
		tmpl []string
		want path.Paths
	}{
		{
			name: "quote",
			tmpl: []string{
				`{{ quote .Values.app.name }}`,
				`{{ .Values.app.name | quote }}`,
			},
			want: path.Paths{path.NewPath("app", "name")},
		},
		{
			name: "upper",
			tmpl: []string{
				`{{ upper .Values.ns }}`,
				`{{ .Values.ns | upper }}`,
			},
			want: path.Paths{path.NewPath("ns")},
		},
		{
			name: "lower",
			tmpl: []string{
				`{{ lower .Values.Kind }}`,
				`{{ .Values.Kind | lower }}`,
			},
			want: path.Paths{path.NewPath("Kind")},
		},
		{
			name: "default literal + value",
			tmpl: []string{
				`{{ default "x" .Values.cfg.path }}`,
				`{{ "x" | default .Values.cfg.path }}`,
			},
			want: path.Paths{path.NewPath("cfg", "path")},
		},
	}

	for _, tc := range cases {
		for i, tmpl := range tc.tmpl {
			name := tc.name
			if i == 1 {
				name += "_piped"
			}
			s.Run(name, func() {
				got, err := parseFile(name+".tmpl", []byte(tmpl))
				s.Require().NoError(err)
				path.EqualPaths(s, tc.want, got)
			})
		}
	}
}

// get/index merge their arguments and return them.
func (s *Unittest) TestParseCommand_Return() {
	cases := []struct {
		name string
		tmpl []string
		want path.Paths
	}{
		{
			name: "get wrapper",
			tmpl: []string{
				`{{ get .Values.app "name"}}`,
				`{{ "name" | get .Values.app }}`,
			},
			want: path.Paths{path.NewPath("app", "name")},
		},
		{
			name: "index one",
			tmpl: []string{
				`{{ index .Values.cfg.path "firstIndex" }}`,
				`{{ "firstIndex" | index index .Values.cfg.path }}`,
			},
			want: path.Paths{path.NewPath("cfg", "path", "firstIndex")},
		},
		{
			name: "index two",
			tmpl: []string{
				`{{ index .Values.cfg.path "firstIndex" "secondIndex" }}`,
				`{{ "secondIndex" | index .Values.cfg.path "firstIndex" }}`,
			},
			want: path.Paths{path.NewPath("cfg", "path", "firstIndex", "secondIndex")},
		},
	}

	for _, tc := range cases {
		for i, tmpl := range tc.tmpl {
			name := tc.name
			if i == 1 {
				name += "_piped"
			}
			s.Run(name, func() {
				got, err := parseFile(tc.name+".tmpl", []byte(tmpl))
				s.Require().NoError(err)
				path.EqualPaths(s, tc.want, got)
			})
		}
	}
}

// Functions out of scope: include/tpl should be treated as not implemented.
func (s *Unittest) TestParseCommand_NotImplemented() {
	cases := []struct {
		name string
		tmpl string
	}{
		{name: "include function", tmpl: `{{ include "x" . }}`},
		{name: "tpl function", tmpl: `{{ tpl "{{ .Values.a.x }}" . }}`},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.Require().Panics(func() {
				_, _ = parseFile(tc.name+".tmpl", []byte(tc.tmpl))
			})
		})
	}
}
