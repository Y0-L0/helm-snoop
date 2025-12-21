package parser

import "github.com/y0-l0/helm-snoop/pkg/path"

// Happy-path tests for parseFile/analyzer.collect.
func (s *Unittest) TestParseFile_Happy() {
	cases := []struct {
		name string
		tmpl string
		want path.Paths
	}{
		{
			name: "two values in order",
			tmpl: `{{ .Values.config.message }} {{ .Values.config.enabled }}`,
			want: path.Paths{path.NewPath("config", "message"), path.NewPath("config", "enabled")},
		},
		{
			name: "ignores bare Values but keeps proper field",
			tmpl: `{{ .Values }} {{ .Values.config.message }}`,
			want: path.Paths{path.NewPath("config", "message")},
		},
		{
			name: "duplicates are preserved",
			tmpl: `{{ .Values.a.b }} {{ .Values.a.b }}`,
			want: path.Paths{path.NewPath("a", "b"), path.NewPath("a", "b")},
		},
		{
			name: "multiline with spaces",
			tmpl: "A: {{    .Values.x.y    }}\nB: {{ .Values.z }}",
			want: path.Paths{path.NewPath("x", "y"), path.NewPath("z")},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			got, err := parseFile(tc.name+".tmpl", []byte(tc.tmpl))
			s.Require().NoError(err)
			path.EqualPaths(s, tc.want, got)
		})
	}
}

// Unhappy-path: control structures are not implemented and should panic
// when encountered by analyzer.collect (for now).
func (s *Unittest) TestParseFile_NotImplemented() {
	cases := []struct {
		name string
		tmpl string
	}{
		{name: "with block", tmpl: `{{ with . }}{{ .Values.a.x }}{{ end }}`},
		{name: "range block", tmpl: `{{ range .Values.items }}{{ end }}`},
		{name: "template action", tmpl: `{{ template "x" . }}`},
	}
	for _, testCase := range cases {
		s.Run(testCase.name, func() {
			s.Require().Panics(func() {
				_, _ = parseFile(testCase.name+".tmpl", []byte(testCase.tmpl))
			})
		})
	}
}

// if/else should be supported: evaluate the condition and both branches.
func (s *Unittest) TestParseFile_IfElse() {
	cases := []struct {
		name string
		tmpl string
		want path.Paths
	}{
		{
			name: "if condition only",
			tmpl: `{{ if .Values.a.x }}ok{{ end }}`,
			want: path.Paths{path.NewPath("a", "x")},
		},
		{
			name: "if with else, values in both",
			tmpl: `{{ if .Values.a.x }}{{ .Values.p }}{{ else }}{{ .Values.q }}{{ end }}`,
			want: path.Paths{path.NewPath("a", "x"), path.NewPath("p"), path.NewPath("q")},
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			got, err := parseFile(tc.name+".tmpl", []byte(tc.tmpl))
			s.Require().NoError(err)
			path.EqualPaths(s, tc.want, got)
		})
	}
}
