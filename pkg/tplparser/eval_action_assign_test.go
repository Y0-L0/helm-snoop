package tplparser

import (
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// Stage 2: Track variable assignments in evalActionNode.
// When a pipe has Decl ($ctx := ...), store the evalResult so downstream
// variable references can recover dict structure.

// Basic: $ctx := dict → include with $ctx → helper resolves dict keys.
func (s *Unittest) TestActionAssign_DictToInclude() {
	paths := s.parseChart(
		testFile{
			"templates/_helpers.yaml",
			`{{ define "test.helper" }}{{ .key.sub }}{{ end }}`,
		},
		testFile{
			"templates/main.yaml",
			`{{ $ctx := dict "key" .Values.foo }}{{ include "test.helper" $ctx }}`,
		},
	)

	expected := vpath.Paths{vpath.NewPath("foo", "sub")}
	vpath.EqualPaths(s, expected, paths)
}

// $ctx := merge(dict, dict) → include with $ctx → helper resolves dict keys.
func (s *Unittest) TestActionAssign_MergeToInclude() {
	paths := s.parseChart(
		testFile{
			"templates/_helpers.yaml",
			`{{ define "test.helper" }}{{ .podSC | toYaml }}{{ end }}`,
		},
		testFile{
			"templates/main.yaml",
			`{{ $ctx := merge (dict "podSC" .Values.secCtx) (dict "extra" "x") }}{{ include "test.helper" $ctx }}`,
		},
	)

	expected := vpath.Paths{
		np().Key("secCtx").Wildcard(),
	}
	vpath.EqualPaths(s, expected, paths)
}

// Real-world pattern: two-layer include with merge assigned to $ctx.
func (s *Unittest) TestActionAssign_TwoLayerMergeInclude() {
	paths := s.parseChart(
		testFile{
			"templates/_inner.yaml",
			`{{ define "inner.helper" }}{{ .podSC | toYaml }}{{ end }}`,
		},
		testFile{
			"templates/_outer.yaml",
			`{{ define "outer.helper" }}{{ $ctx := merge (dict "podSC" (omit .securityContext "enabled")) (dict "extra" "x") }}{{ include "inner.helper" $ctx }}{{ end }}`,
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

// Assigned variable should not leak into included templates.
func (s *Unittest) TestActionAssign_VariableDoesNotLeakIntoIncludes() {
	paths := s.parseChart(
		testFile{
			"templates/_helpers.yaml",
			// Helper tries to use $ctx — should not resolve
			`{{ define "test.helper" }}{{ .Values.direct }}{{ end }}`,
		},
		testFile{
			"templates/main.yaml",
			`{{ $ctx := dict "key" .Values.foo }}{{ include "test.helper" . }}`,
		},
	)

	// Only .direct should be found — $ctx is not passed to the include
	expected := vpath.Paths{vpath.NewPath("direct")}
	vpath.EqualPaths(s, expected, paths)
}

// Assigned variable with plain path (no dict) should still work.
// The path itself is emitted as context, plus field access inside the helper.
func (s *Unittest) TestActionAssign_PlainPathVariable() {
	paths := s.parseChart(
		testFile{
			"templates/_helpers.yaml",
			`{{ define "test.helper" }}{{ .sub }}{{ end }}`,
		},
		testFile{
			"templates/main.yaml",
			`{{ $ctx := .Values.config }}{{ include "test.helper" $ctx }}`,
		},
	)

	expected := vpath.Paths{
		vpath.NewPath("config"),
		vpath.NewPath("config", "sub"),
	}
	vpath.EqualPaths(s, expected, paths)
}

// Variable used directly (not via include) — field access on assigned dict.
func (s *Unittest) TestActionAssign_DirectFieldAccess() {
	tmpl := `{{ $ctx := dict "key" .Values.foo }}{{ $ctx.key }}`

	paths, err := parseFile("", "test.yaml", []byte(tmpl), nil)
	s.Require().NoError(err)

	// $ctx.key should resolve through the dict to .Values.foo
	expected := vpath.Paths{vpath.NewPath("foo")}
	vpath.EqualPaths(s, expected, paths)
}

// Variable reassignment: second assignment should override.
func (s *Unittest) TestActionAssign_Reassignment() {
	paths := s.parseChart(
		testFile{
			"templates/_helpers.yaml",
			`{{ define "test.helper" }}{{ .key.sub }}{{ end }}`,
		},
		testFile{
			"templates/main.yaml",
			`{{ $ctx := dict "key" .Values.first }}{{ $ctx = dict "key" .Values.second }}{{ include "test.helper" $ctx }}`,
		},
	)

	// Second assignment wins
	expected := vpath.Paths{vpath.NewPath("second", "sub")}
	vpath.EqualPaths(s, expected, paths)
}

// Variable bound to root ($) with .Values access strips the Values segment.
func (s *Unittest) TestActionAssign_RootBoundVariableStripsValues() {
	paths := s.parseChart(
		testFile{
			"templates/_helpers.yaml",
			`{{ define "test.helper" }}{{ .Values.global.imagePullSecrets }}{{ end }}`,
		},
		testFile{
			"templates/main.yaml",
			`{{ $context := . }}{{ include "test.helper" $context }}`,
		},
	)

	expected := vpath.Paths{vpath.NewPath("global", "imagePullSecrets")}
	vpath.EqualPaths(s, expected, paths)
}
