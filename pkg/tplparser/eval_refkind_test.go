package tplparser

import "github.com/y0-l0/helm-snoop/pkg/vpath"

func (s *Unittest) TestRefKind_IfConditionMarkedChecked() {
	paths := s.parse(`{{ if .Values.enabled }}yes{{ end }}`)

	s.Require().Len(paths, 1)
	vpath.EqualPath(s, vpath.NewCheckedPath("enabled"), paths[0])
}

func (s *Unittest) TestRefKind_IfBodyRemainsConsumed() {
	paths := s.parse(`{{ if .Values.enabled }}{{ .Values.name }}{{ end }}`)

	s.Require().Len(paths, 2)
	vpath.EqualPath(s, vpath.NewCheckedPath("enabled"), paths[0])
	vpath.EqualPath(s, vpath.NewPath("name"), paths[1])
}

func (s *Unittest) TestRefKind_ElseIfConditionMarkedChecked() {
	paths := s.parse(`{{ if .Values.a }}x{{ else if .Values.b }}y{{ end }}`)

	s.Require().Len(paths, 2)
	vpath.EqualPath(s, vpath.NewCheckedPath("a"), paths[0])
	vpath.EqualPath(s, vpath.NewCheckedPath("b"), paths[1])
}

func (s *Unittest) TestRefKind_AndArgsMarkedChecked() {
	paths := s.parse(`{{ if and .Values.a .Values.a.name }}yes{{ end }}`)

	s.Require().Len(paths, 2)
	vpath.EqualPath(s, vpath.NewCheckedPath("a"), paths[0])
	vpath.EqualPath(s, vpath.NewCheckedPath("a", "name"), paths[1])
}

func (s *Unittest) TestRefKind_OrArgsMarkedChecked() {
	paths := s.parse(`{{ if or .Values.a .Values.b }}yes{{ end }}`)

	s.Require().Len(paths, 2)
	vpath.EqualPath(s, vpath.NewCheckedPath("a"), paths[0])
	vpath.EqualPath(s, vpath.NewCheckedPath("b"), paths[1])
}

func (s *Unittest) TestRefKind_LenMarkedChecked() {
	paths := s.parse(`{{ len .Values.items }}`)

	s.Require().Len(paths, 1)
	vpath.EqualPath(s, vpath.NewCheckedPath("items"), paths[0])
}

func (s *Unittest) TestRefKind_ListMarkedChecked() {
	paths := s.parse(`{{ list .Values.a }}`)

	s.Require().Len(paths, 1)
	vpath.EqualPath(s, vpath.NewCheckedPath("a"), paths[0])
}

func (s *Unittest) TestRefKind_FirstMarkedChecked() {
	paths := s.parse(`{{ first .Values.items }}`)

	s.Require().Len(paths, 1)
	vpath.EqualPath(s, vpath.NewCheckedPath("items"), paths[0])
}

func (s *Unittest) TestRefKind_LastMarkedChecked() {
	paths := s.parse(`{{ last .Values.items }}`)

	s.Require().Len(paths, 1)
	vpath.EqualPath(s, vpath.NewCheckedPath("items"), paths[0])
}

func (s *Unittest) TestRefKind_AndBodyRemainsConsumed() {
	paths := s.parse(`{{ if and .Values.a .Values.b }}{{ .Values.name }}{{ end }}`)

	s.Require().Len(paths, 3)
	vpath.EqualPath(s, vpath.NewCheckedPath("a"), paths[0])
	vpath.EqualPath(s, vpath.NewCheckedPath("b"), paths[1])
	vpath.EqualPath(s, vpath.NewPath("name"), paths[2])
}
