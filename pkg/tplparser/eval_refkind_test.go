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

func (s *Unittest) TestRefKind_AndArgsMarkedChecked() {
	paths := s.parse(`{{ if and .Values.a .Values.a.name }}yes{{ end }}`)

	s.Require().Len(paths, 2)
	s.Equal(".a", paths[0].ID())
	s.Equal(vpath.Checked, paths[0].Usage)
	s.Equal(".a.name", paths[1].ID())
	s.Equal(vpath.Checked, paths[1].Usage)
}

func (s *Unittest) TestRefKind_OrArgsMarkedChecked() {
	paths := s.parse(`{{ if or .Values.a .Values.b }}yes{{ end }}`)

	s.Require().Len(paths, 2)
	s.Equal(".a", paths[0].ID())
	s.Equal(vpath.Checked, paths[0].Usage)
	s.Equal(".b", paths[1].ID())
	s.Equal(vpath.Checked, paths[1].Usage)
}

func (s *Unittest) TestRefKind_LenMarkedChecked() {
	paths := s.parse(`{{ len .Values.items }}`)

	s.Require().Len(paths, 1)
	s.Equal(".items", paths[0].ID())
	s.Equal(vpath.Checked, paths[0].Usage)
}

func (s *Unittest) TestRefKind_ListMarkedChecked() {
	paths := s.parse(`{{ list .Values.a }}`)

	s.Require().Len(paths, 1)
	s.Equal(".a", paths[0].ID())
	s.Equal(vpath.Checked, paths[0].Usage)
}

func (s *Unittest) TestRefKind_AndBodyRemainsConsumed() {
	paths := s.parse(`{{ if and .Values.a .Values.b }}{{ .Values.name }}{{ end }}`)

	s.Require().Len(paths, 3)
	s.Equal(".a", paths[0].ID())
	s.Equal(vpath.Checked, paths[0].Usage)
	s.Equal(".b", paths[1].ID())
	s.Equal(vpath.Checked, paths[1].Usage)
	s.Equal(".name", paths[2].ID())
	s.Equal(vpath.Consumed, paths[2].Usage)
}
