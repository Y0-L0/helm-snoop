package tplparser

import "github.com/y0-l0/helm-snoop/pkg/vpath"

// NoValues: template without any .Values usage should yield empty slice.
func (s *Unittest) TestParseFile_NoValues() {
	tmpl := `kind: ConfigMap\nmetadata: { name: test }\n# no values here\nliteral: text`
	got, err := parseFile("", "novals.tmpl", []byte(tmpl), nil)
	s.Require().NoError(err)
	s.Require().Empty(got)
}

// TestGetUsages_CalledDefinesAreNotDoubleEvaluated checks that a define which
// IS called via include is not evaluated a second time as an uncalled define.
func (s *Unittest) TestGetUsages_CalledDefinesAreNotDoubleEvaluated() {
	c := buildChart(
		testFile{"templates/_helpers.tpl", `{{- define "myhelper" -}}{{ .Values.foo.bar }}{{- end -}}`},
		testFile{"templates/cm.yaml", `data: { a: {{ include "myhelper" . }} }`},
	)
	paths, err := GetUsages(c)
	s.Require().NoError(err)
	s.Require().Len(paths, 1)
	vpath.EqualPath(s, vpath.NewPath("foo", "bar"), paths[0])
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
