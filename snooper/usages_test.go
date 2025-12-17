package snooper

// NoValues: template without any .Values usage should yield empty slice.
func (s *Unittest) TestParseFile_NoValues() {
	tmpl := `kind: ConfigMap\nmetadata: { name: test }\n# no values here\nliteral: text`
	got, err := parseFile("novals.tmpl", []byte(tmpl))
	s.Require().NoError(err)
	s.Require().Empty(got)
}

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
		// {name: "tpl function", tmpl: `{{ tpl "{{ .Values.a.x }}" . }}`},
	}
	for _, testCase := range cases {
		s.Run(testCase.name, func() {
			s.Require().Panics(func() {
				_, _ = parseFile(testCase.name+".tmpl", []byte(testCase.tmpl))
			})
		})
	}
}

// Invalid template syntax should return an error, not panic.
func (s *Unittest) TestParseFile_InvalidTemplate() {
	cases := []string{
		"{{",                    // unclosed action
		`{{ .Values.config. }}`, // invalid field
		`{{ if }}`,              // invalid if syntax
	}
	for i, src := range cases {
		s.Run("invalid-"+string(rune('a'+i)), func() {
			_, err := parseFile("invalid.tmpl", []byte(src))
			s.Require().Error(err)
		})
	}
}
