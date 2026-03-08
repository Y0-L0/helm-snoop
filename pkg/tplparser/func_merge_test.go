package tplparser

import (
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// merge should preserve dict structure so include can resolve params.
func (s *Unittest) TestMerge_PreservesDictStructure() {
	paths := s.parseChart(
		testFile{"templates/_helpers.yaml", `{{ define "test.helper" }}{{ .key.sub }}{{ end }}`},
		testFile{
			"templates/main.yaml",
			`{{ include "test.helper" (merge (dict "key" .Values.foo) (dict "extra" "lit")) }}`,
		},
	)

	expected := vpath.Paths{vpath.NewPath("foo", "sub")}
	vpath.EqualPaths(s, expected, paths)
}

// merge follows Sprig semantics: first dict wins for duplicate keys.
func (s *Unittest) TestMerge_FirstDictWins() {
	paths := s.parseChart(
		testFile{"templates/_helpers.yaml", `{{ define "test.helper" }}{{ .val.sub }}{{ end }}`},
		testFile{
			"templates/main.yaml",
			`{{ include "test.helper" (merge (dict "val" .Values.first) (dict "val" .Values.second)) }}`,
		},
	)

	// First dict wins: .val.sub resolves to .first.sub
	expected := vpath.Paths{vpath.NewPath("first", "sub")}
	vpath.EqualPaths(s, expected, paths)
}

// mergeOverwrite follows Sprig semantics: last dict wins for duplicate keys.
func (s *Unittest) TestMergeOverwrite_LastDictWins() {
	paths := s.parseChart(
		testFile{"templates/_helpers.yaml", `{{ define "test.helper" }}{{ .val.sub }}{{ end }}`},
		testFile{
			"templates/main.yaml",
			`{{ include "test.helper" (mergeOverwrite (dict "val" .Values.first) (dict "val" .Values.second)) }}`,
		},
	)

	// Last dict wins: .val.sub resolves to .second.sub
	expected := vpath.Paths{vpath.NewPath("second", "sub")}
	vpath.EqualPaths(s, expected, paths)
}

// merge with non-dict args should still emit paths (backwards compatible).
func (s *Unittest) TestMerge_NonDictArgsEmitPaths() {
	tmpl := `{{ merge .Values.a .Values.b }}`

	paths, err := parseFile("", "test.yaml", []byte(tmpl), nil)
	s.Require().NoError(err)

	expected := vpath.Paths{
		vpath.NewPath("a"),
		vpath.NewPath("b"),
	}
	vpath.EqualPaths(s, expected, paths)
}

// Reproducer for the real-world bug: two-layer include with merge in between.
// rmqco helper merges a synthetic dict and passes it to a cloudpirates helper.
func (s *Unittest) TestMerge_TwoLayerIncludeWithSyntheticContext() {
	paths := s.parseChart(
		testFile{"templates/_inner.yaml", `{{ define "inner.helper" }}{{ .podSC | toYaml }}{{ end }}`},
		testFile{
			"templates/_outer.yaml",
			`{{ define "outer.helper" }}{{ include "inner.helper" (merge (dict "podSC" (omit .securityContext "enabled")) (dict "extra" "x")) }}{{ end }}`,
		},
		testFile{
			"templates/main.yaml",
			`{{ include "outer.helper" (dict "securityContext" .Values.pod.securityContext) }}`,
		},
	)

	expected := vpath.Paths{
		np().Key("pod").Key("securityContext").Wildcard(),
	}
	vpath.EqualPaths(s, expected, paths)
}
