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
				got, err := parseFile("", name+".tmpl", []byte(tmpl), nil)
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
			want: path.Paths{path.NewPath("app").Any("name")},
		},
		{
			name: "index one",
			tmpl: []string{
				`{{ index .Values.cfg.path "firstIndex" }}`,
				`{{ "firstIndex" | index .Values.cfg.path }}`,
			},
			want: path.Paths{path.NewPath("cfg", "path").Any("firstIndex")},
		},
		{
			name: "index two",
			tmpl: []string{
				`{{ index .Values.cfg.path "firstIndex" "secondIndex" }}`,
				`{{ "secondIndex" | index .Values.cfg.path "firstIndex" }}`,
			},
			want: path.Paths{path.NewPath("cfg", "path").Any("firstIndex").Any("secondIndex")},
		},
	}

	for _, tc := range cases {
		for i, tmpl := range tc.tmpl {
			name := tc.name
			if i == 1 {
				name += "_piped"
			}
			s.Run(name, func() {
				got, err := parseFile("", tc.name+".tmpl", []byte(tmpl), nil)
				s.Require().NoError(err)
				path.EqualPaths(s, tc.want, got)
			})
		}
	}
}

// Test complex function calls (with multiple args) being piped to other functions.
// This tests that non-piped function calls can be piped as a whole.
func (s *Unittest) TestParseCommand_ComplexPipe() {
	cases := []struct {
		name string
		tmpl string
		want path.Paths
	}{
		{
			name: "index then quote",
			tmpl: `{{ index .Values.cfg.path "firstIndex" | quote }}`,
			want: path.Paths{path.NewPath("cfg", "path").Any("firstIndex")},
		},
		{
			name: "get then upper",
			tmpl: `{{ get .Values.app "name" | upper }}`,
			want: path.Paths{path.NewPath("app").Any("name")},
		},
		{
			name: "default then lower",
			tmpl: `{{ default "fallback" .Values.config | lower }}`,
			want: path.Paths{path.NewPath("config")},
		},
		{
			name: "index two keys then quote",
			tmpl: `{{ index .Values.nested "key1" "key2" | quote }}`,
			want: path.Paths{path.NewPath("nested").Any("key1").Any("key2")},
		},
		{
			name: "triple pipe",
			tmpl: `{{ .Values.data | quote | upper | lower }}`,
			want: path.Paths{path.NewPath("data")},
		},
		{
			name: "complex triple pipe",
			tmpl: `{{ index .Values.foo "bar" | quote | upper }}`,
			want: path.Paths{path.NewPath("foo").Any("bar")},
		},
		{
			name: "default chain preserves all",
			tmpl: `{{ .Values.a | default .Values.b | default .Values.c }}`,
			want: path.Paths{path.NewPath("a"), path.NewPath("b"), path.NewPath("c")},
		},
		{
			name: "default with piped value",
			tmpl: `{{ .Values.primary | default .Values.fallback }}`,
			want: path.Paths{path.NewPath("primary"), path.NewPath("fallback")},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			got, err := parseFile("", tc.name+".tmpl", []byte(tc.tmpl), nil)
			s.Require().NoError(err)
			path.EqualPaths(s, tc.want, got)
		})
	}
}
