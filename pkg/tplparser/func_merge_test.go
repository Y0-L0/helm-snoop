package tplparser

import (
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"

	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// merge should preserve dict structure so include can resolve params.
func (s *Unittest) TestMerge_PreservesDictStructure() {
	helperTpl := `{{ define "test.helper" }}{{ .key.sub }}{{ end }}`
	mainTpl := `{{ include "test.helper" (merge (dict "key" .Values.foo) (dict "extra" "lit")) }}`

	c := &chart.Chart{Templates: []*common.File{
		{Name: "templates/_helpers.yaml", Data: []byte(helperTpl)},
		{Name: "templates/main.yaml", Data: []byte(mainTpl)},
	}}

	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	paths, err := parseFile("", "templates/main.yaml", c.Templates[1].Data, idx)
	s.Require().NoError(err)

	// .key.sub in the helper should resolve to .foo.sub via paramPaths
	expected := vpath.Paths{
		vpath.NewPath("foo", "sub"),
	}
	vpath.EqualPaths(s, expected, paths)
}

// merge follows Sprig semantics: first dict wins for duplicate keys.
func (s *Unittest) TestMerge_FirstDictWins() {
	helperTpl := `{{ define "test.helper" }}{{ .val.sub }}{{ end }}`
	mainTpl := `{{ include "test.helper" (merge (dict "val" .Values.first) (dict "val" .Values.second)) }}`

	c := &chart.Chart{Templates: []*common.File{
		{Name: "templates/_helpers.yaml", Data: []byte(helperTpl)},
		{Name: "templates/main.yaml", Data: []byte(mainTpl)},
	}}

	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	paths, err := parseFile("", "templates/main.yaml", c.Templates[1].Data, idx)
	s.Require().NoError(err)

	// First dict wins: .val.sub resolves to .first.sub
	// .second is overridden and not emitted (dict values only emit via param resolution)
	expected := vpath.Paths{
		vpath.NewPath("first", "sub"),
	}
	vpath.EqualPaths(s, expected, paths)
}

// mergeOverwrite follows Sprig semantics: last dict wins for duplicate keys.
func (s *Unittest) TestMergeOverwrite_LastDictWins() {
	helperTpl := `{{ define "test.helper" }}{{ .val.sub }}{{ end }}`
	mainTpl := `{{ include "test.helper" (mergeOverwrite (dict "val" .Values.first) (dict "val" .Values.second)) }}`

	c := &chart.Chart{Templates: []*common.File{
		{Name: "templates/_helpers.yaml", Data: []byte(helperTpl)},
		{Name: "templates/main.yaml", Data: []byte(mainTpl)},
	}}

	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	paths, err := parseFile("", "templates/main.yaml", c.Templates[1].Data, idx)
	s.Require().NoError(err)

	// Last dict wins: .val.sub resolves to .second.sub
	// .first is overridden and not emitted (dict values only emit via param resolution)
	expected := vpath.Paths{
		vpath.NewPath("second", "sub"),
	}
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
	// Inner helper: accesses .podSC from dict params, serializes it
	innerTpl := `{{ define "inner.helper" }}{{ .podSC | toYaml }}{{ end }}`
	// Outer helper: merges dict with omit result, passes to inner
	outerTpl := `{{ define "outer.helper" }}{{ include "inner.helper" (merge (dict "podSC" (omit .securityContext "enabled")) (dict "extra" "x")) }}{{ end }}`
	// Main template: passes whole-object path to outer helper via dict
	mainTpl := `{{ include "outer.helper" (dict "securityContext" .Values.pod.securityContext) }}`

	c := &chart.Chart{Templates: []*common.File{
		{Name: "templates/_inner.yaml", Data: []byte(innerTpl)},
		{Name: "templates/_outer.yaml", Data: []byte(outerTpl)},
		{Name: "templates/main.yaml", Data: []byte(mainTpl)},
	}}

	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	paths, err := parseFile("", "templates/main.yaml", c.Templates[2].Data, idx)
	s.Require().NoError(err)

	// The path .pod.securityContext should flow through:
	// 1. dict "securityContext" .Values.pod.securityContext → paramPaths[securityContext] = .pod.securityContext
	// 2. omit .securityContext "enabled" → returns path .pod.securityContext
	// 3. merge (dict "podSC" <above>) ... → should preserve dict with podSC = .pod.securityContext
	// 4. include "inner.helper" <merged> → paramPaths[podSC] = .pod.securityContext
	// 5. .podSC | toYaml → emits .pod.securityContext.*
	expected := vpath.Paths{
		np().Key("pod").Key("securityContext").Wildcard(),
	}
	vpath.EqualPaths(s, expected, paths)
}
