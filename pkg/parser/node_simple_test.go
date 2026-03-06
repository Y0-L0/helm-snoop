package parser

import "github.com/y0-l0/helm-snoop/pkg/vpath"

// quote/upper/lower simply wrap a value; default X Y reads only Y when X is literal.
func (s *Unittest) TestParseCommand_Noops() {
	cases := []struct {
		name string
		tmpl []string
		want vpath.Paths
	}{
		{
			name: "quote",
			tmpl: []string{
				`{{ quote .Values.app.name }}`,
				`{{ .Values.app.name | quote }}`,
			},
			want: vpath.Paths{vpath.NewPath("app", "name")},
		},
		{
			name: "upper",
			tmpl: []string{
				`{{ upper .Values.ns }}`,
				`{{ .Values.ns | upper }}`,
			},
			want: vpath.Paths{vpath.NewPath("ns")},
		},
		{
			name: "lower",
			tmpl: []string{
				`{{ lower .Values.Kind }}`,
				`{{ .Values.Kind | lower }}`,
			},
			want: vpath.Paths{vpath.NewPath("Kind")},
		},
		{
			name: "default literal + value",
			tmpl: []string{
				`{{ default "x" .Values.cfg.path }}`,
				`{{ "x" | default .Values.cfg.path }}`,
			},
			want: vpath.Paths{vpath.NewPath("cfg", "path")},
		},
	}

	for _, tc := range cases {
		for i, tmpl := range tc.tmpl {
			name := tc.name
			if i == 1 {
				name += "_piped"
			}
			s.Run(name, func() {
				vpath.EqualPaths(s, tc.want, s.parse(tmpl))
			})
		}
	}
}

// get/index merge their arguments and return them.
func (s *Unittest) TestParseCommand_Return() {
	cases := []struct {
		name string
		tmpl []string
		want vpath.Paths
	}{
		{
			name: "get wrapper",
			tmpl: []string{
				`{{ get .Values.app "name"}}`,
				`{{ "name" | get .Values.app }}`,
			},
			want: vpath.Paths{vpath.NewPath("app").Any("name")},
		},
		{
			name: "index one",
			tmpl: []string{
				`{{ index .Values.cfg.path "firstIndex" }}`,
				`{{ "firstIndex" | index .Values.cfg.path }}`,
			},
			want: vpath.Paths{vpath.NewPath("cfg", "path").Any("firstIndex")},
		},
		{
			name: "index two",
			tmpl: []string{
				`{{ index .Values.cfg.path "firstIndex" "secondIndex" }}`,
				`{{ "secondIndex" | index .Values.cfg.path "firstIndex" }}`,
			},
			want: vpath.Paths{vpath.NewPath("cfg", "path").Any("firstIndex").Any("secondIndex")},
		},
	}

	for _, tc := range cases {
		for i, tmpl := range tc.tmpl {
			name := tc.name
			if i == 1 {
				name += "_piped"
			}
			s.Run(name, func() {
				vpath.EqualPaths(s, tc.want, s.parse(tmpl))
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
		want vpath.Paths
	}{
		{
			name: "index then quote",
			tmpl: `{{ index .Values.cfg.path "firstIndex" | quote }}`,
			want: vpath.Paths{vpath.NewPath("cfg", "path").Any("firstIndex")},
		},
		{
			name: "get then upper",
			tmpl: `{{ get .Values.app "name" | upper }}`,
			want: vpath.Paths{vpath.NewPath("app").Any("name")},
		},
		{
			name: "default then lower",
			tmpl: `{{ default "fallback" .Values.config | lower }}`,
			want: vpath.Paths{vpath.NewPath("config")},
		},
		{
			name: "index two keys then quote",
			tmpl: `{{ index .Values.nested "key1" "key2" | quote }}`,
			want: vpath.Paths{vpath.NewPath("nested").Any("key1").Any("key2")},
		},
		{
			name: "triple pipe",
			tmpl: `{{ .Values.data | quote | upper | lower }}`,
			want: vpath.Paths{vpath.NewPath("data")},
		},
		{
			name: "complex triple pipe",
			tmpl: `{{ index .Values.foo "bar" | quote | upper }}`,
			want: vpath.Paths{vpath.NewPath("foo").Any("bar")},
		},
		{
			name: "default chain preserves all",
			tmpl: `{{ .Values.a | default .Values.b | default .Values.c }}`,
			want: vpath.Paths{vpath.NewPath("a"), vpath.NewPath("b"), vpath.NewPath("c")},
		},
		{
			name: "default with piped value",
			tmpl: `{{ .Values.primary | default .Values.fallback }}`,
			want: vpath.Paths{vpath.NewPath("primary"), vpath.NewPath("fallback")},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			vpath.EqualPaths(s, tc.want, s.parse(tc.tmpl))
		})
	}
}
