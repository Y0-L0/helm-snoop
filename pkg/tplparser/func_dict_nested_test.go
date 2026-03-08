package tplparser

import (
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// parseNestedDict builds a chart with a single helper and a main template that includes it.
func (s *Unittest) parseNestedDict(helperBody, mainBody string) vpath.Paths {
	return s.parseChart(
		testFile{"templates/_helpers.yaml", `{{ define "test.helper" }}` + helperBody + `{{ end }}`},
		testFile{"templates/main.yaml", mainBody},
	)
}

func (s *Unittest) TestNestedDict_TwoLevelNesting() {
	paths := s.parseNestedDict(
		`{{ .outer.inner.sub }}`,
		`{{ include "test.helper" (dict "outer" (dict "inner" .Values.foo)) }}`,
	)
	vpath.EqualPaths(s, vpath.Paths{vpath.NewPath("foo", "sub")}, paths)
}

func (s *Unittest) TestNestedDict_ValuesWrapper() {
	paths := s.parseNestedDict(
		`{{ .Values.podSC | toYaml }}`,
		`{{ include "test.helper" (dict "Values" (dict "podSC" .Values.pod.securityContext)) }}`,
	)
	vpath.EqualPaths(s, vpath.Paths{np().Key("pod").Key("securityContext").Wildcard()}, paths)
}

func (s *Unittest) TestNestedDict_ThreeLevels() {
	paths := s.parseNestedDict(
		`{{ .a.b.c.sub }}`,
		`{{ include "test.helper" (dict "a" (dict "b" (dict "c" .Values.x))) }}`,
	)
	vpath.EqualPaths(s, vpath.Paths{vpath.NewPath("x", "sub")}, paths)
}

func (s *Unittest) TestNestedDict_MixedPathAndDict() {
	paths := s.parseNestedDict(
		`{{ .plain.sub }}{{ .nested.key.sub }}`,
		`{{ include "test.helper" (dict "plain" .Values.a "nested" (dict "key" .Values.b)) }}`,
	)
	expected := vpath.Paths{
		vpath.NewPath("a", "sub"),
		vpath.NewPath("b", "sub"),
	}
	vpath.EqualPaths(s, expected, paths)
}

func (s *Unittest) TestNestedDict_IndexReturnsInnerDict() {
	paths := s.parseNestedDict(
		`{{ .b.sub }}`,
		`{{ include "test.helper" (index (dict "a" (dict "b" .Values.x)) "a") }}`,
	)
	vpath.EqualPaths(s, vpath.Paths{vpath.NewPath("x", "sub")}, paths)
}

func (s *Unittest) TestNestedDict_GetReturnsInnerDict() {
	paths := s.parseNestedDict(
		`{{ .b.sub }}`,
		`{{ include "test.helper" (get (dict "a" (dict "b" .Values.x)) "a") }}`,
	)
	vpath.EqualPaths(s, vpath.Paths{vpath.NewPath("x", "sub")}, paths)
}

// Two-layer include with merge — needs three template files.
func (s *Unittest) TestNestedDict_MergeWithValuesWrapper() {
	paths := s.parseChart(
		testFile{
			"templates/_helpers.yaml",
			`{{ define "test.helper" }}{{ .Values.podSC | toYaml }}{{ end }}`,
		},
		testFile{
			"templates/_outer.yaml",
			`{{ define "outer.helper" }}{{ include "test.helper" (merge (dict "Values" (dict "podSC" (omit .securityContext "enabled"))) .context) }}{{ end }}`,
		},
		testFile{
			"templates/main.yaml",
			`{{ include "outer.helper" (dict "securityContext" .Values.pod.securityContext "context" .) }}`,
		},
	)
	vpath.EqualPaths(s, vpath.Paths{np().Key("pod").Key("securityContext").Wildcard()}, paths)
}
