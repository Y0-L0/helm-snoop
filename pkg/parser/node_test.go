package parser

import "github.com/y0-l0/helm-snoop/pkg/vpath"

// Happy-path tests for parseFile/analyzer.collect.
func (s *Unittest) TestParseFile_Happy() {
	cases := []struct {
		name string
		tmpl string
		want vpath.Paths
	}{
		{
			name: "two values in order",
			tmpl: `{{ .Values.config.message }} {{ .Values.config.enabled }}`,
			want: vpath.Paths{vpath.NewPath("config", "message"), vpath.NewPath("config", "enabled")},
		},
		{
			name: "ignores bare Values but keeps proper field",
			tmpl: `{{ .Values }} {{ .Values.config.message }}`,
			want: vpath.Paths{vpath.NewPath("config", "message")},
		},
		{
			name: "duplicates are preserved",
			tmpl: `{{ .Values.a.b }} {{ .Values.a.b }}`,
			want: vpath.Paths{vpath.NewPath("a", "b"), vpath.NewPath("a", "b")},
		},
		{
			name: "multiline with spaces",
			tmpl: "A: {{    .Values.x.y    }}\nB: {{ .Values.z }}",
			want: vpath.Paths{vpath.NewPath("x", "y"), vpath.NewPath("z")},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			vpath.EqualPaths(s, tc.want, s.parse(tc.tmpl))
		})
	}
}

// with/range/template are now supported.
func (s *Unittest) TestParseFile_NotImplemented() {
	// All previously unsupported features are now implemented
	// This test now verifies they don't panic
	cases := []struct {
		name string
		tmpl string
		want vpath.Paths
	}{
		{
			name: "with block",
			tmpl: `{{ with . }}{{ .Values.a.x }}{{ end }}`,
			want: vpath.Paths{vpath.NewPath("a", "x")},
		},
		{name: "range block", tmpl: `{{ range .Values.items }}{{ end }}`, want: vpath.Paths{}},
		{name: "template action", tmpl: `{{ template "x" . }}`, want: vpath.Paths{}},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			vpath.EqualPaths(s, tc.want, s.parse(tc.tmpl))
		})
	}
}

// if/else should be supported: evaluate the condition and both branches.
func (s *Unittest) TestParseFile_IfElse() {
	cases := []struct {
		name string
		tmpl string
		want vpath.Paths
	}{
		{
			name: "if condition only",
			tmpl: `{{ if .Values.a.x }}ok{{ end }}`,
			want: vpath.Paths{vpath.NewPath("a", "x")},
		},
		{
			name: "if with else, values in both",
			tmpl: `{{ if .Values.a.x }}{{ .Values.p }}{{ else }}{{ .Values.q }}{{ end }}`,
			want: vpath.Paths{vpath.NewPath("a", "x"), vpath.NewPath("p"), vpath.NewPath("q")},
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			vpath.EqualPaths(s, tc.want, s.parse(tc.tmpl))
		})
	}
}
