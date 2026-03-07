package tplparser

import "github.com/y0-l0/helm-snoop/pkg/vpath"

func (s *Unittest) TestRefKind_IfConditionMarkedChecked() {
	paths := s.parse(`{{ if .Values.enabled }}yes{{ end }}`)

	s.Require().Len(paths, 1)
	s.Equal(".enabled", paths[0].ID())
	s.Equal(vpath.Checked, paths[0].Usage)
}

func (s *Unittest) TestRefKind_IfBodyRemainsConsumed() {
	paths := s.parse(`{{ if .Values.enabled }}{{ .Values.name }}{{ end }}`)

	s.Require().Len(paths, 2)
	s.Equal(".enabled", paths[0].ID())
	s.Equal(vpath.Checked, paths[0].Usage)
	s.Equal(".name", paths[1].ID())
	s.Equal(vpath.Consumed, paths[1].Usage)
}

func (s *Unittest) TestRefKind_ElseIfConditionMarkedChecked() {
	paths := s.parse(`{{ if .Values.a }}x{{ else if .Values.b }}y{{ end }}`)

	s.Require().Len(paths, 2)
	s.Equal(".a", paths[0].ID())
	s.Equal(vpath.Checked, paths[0].Usage)
	s.Equal(".b", paths[1].ID())
	s.Equal(vpath.Checked, paths[1].Usage)
}
