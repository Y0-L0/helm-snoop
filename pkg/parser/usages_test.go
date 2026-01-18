package parser

// NoValues: template without any .Values usage should yield empty slice.
func (s *Unittest) TestParseFile_NoValues() {
	tmpl := `kind: ConfigMap\nmetadata: { name: test }\n# no values here\nliteral: text`
	got, err := parseFile("", "novals.tmpl", []byte(tmpl), nil)
	s.Require().NoError(err)
	s.Require().Empty(got)
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
			_, err := parseFile("", "invalid.tmpl", []byte(src), nil)
			s.Require().Error(err)
		})
	}
}
