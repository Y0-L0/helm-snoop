package parser

import "github.com/y0-l0/helm-snoop/pkg/path"

func (s *Unittest) TestParseFile_AttachesFileName() {
	tmpl := `{{ .Values.foo }}`
	paths, err := parseFile("templates/test.yaml", []byte(tmpl), nil)
	s.Require().NoError(err)
	s.Require().Len(paths, 1)
	s.Require().Len(paths[0].Contexts, 1)
	s.Equal("templates/test.yaml", paths[0].Contexts[0].FileName)
}

func (s *Unittest) TestParseFile_AttachesLine() {
	tmpl := "line1\n{{ .Values.foo }}"
	paths, err := parseFile("test.yaml", []byte(tmpl), nil)
	s.Require().NoError(err)
	s.Require().Len(paths, 1)
	s.Require().Len(paths[0].Contexts, 1)
	s.Equal(2, paths[0].Contexts[0].Line)
}

func (s *Unittest) TestParseFile_MultiplePathsDifferentLines() {
	tmpl := "{{ .Values.first }}\n{{ .Values.second }}"
	paths, err := parseFile("test.yaml", []byte(tmpl), nil)
	s.Require().NoError(err)
	path.SortDedup(paths)
	s.Require().Len(paths, 2)

	s.Equal("test.yaml", paths[0].Contexts[0].FileName)
	s.Equal(1, paths[0].Contexts[0].Line)

	s.Equal("test.yaml", paths[1].Contexts[0].FileName)
	s.Equal(2, paths[1].Contexts[0].Line)
}
