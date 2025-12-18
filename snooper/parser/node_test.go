package parser

// Happy-path tests for parseFile/collectUsedValues.
func (s *Unittest) TestParseFile_Happy() {
	cases := []struct {
		name string
		tmpl string
		want []string
	}{
		{
			name: "two values in order",
			tmpl: `{{ .Values.config.message }} {{ .Values.config.enabled }}`,
			want: []string{"config.message", "config.enabled"},
		},
		{
			name: "ignores bare Values but keeps proper field",
			tmpl: `{{ .Values }} {{ .Values.config.message }}`,
			want: []string{"config.message"},
		},
		{
			name: "duplicates are preserved",
			tmpl: `{{ .Values.a.b }} {{ .Values.a.b }}`,
			want: []string{"a.b", "a.b"},
		},
		{
			name: "multiline with spaces",
			tmpl: "A: {{    .Values.x.y    }}\nB: {{ .Values.z }}",
			want: []string{"x.y", "z"},
		},
	}

	for _, testCase := range cases {
		s.Run(testCase.name, func() {
			got, err := parseFile(testCase.name+".tmpl", []byte(testCase.tmpl))
			s.Require().NoError(err)
			s.Require().Equal(testCase.want, got)
		})
	}
}

// Unhappy-path: control structures are not implemented and should panic
// when encountered by collectUsedValues (for now).
func (s *Unittest) TestParseFile_NotImplemented() {
	cases := []struct {
		name string
		tmpl string
	}{
		{name: "if block", tmpl: `{{ if .Values.a.x }}ok{{ end }}`},
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
