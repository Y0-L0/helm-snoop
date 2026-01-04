package parser

import (
	"fmt"
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

// include "tpl.a" . should traverse the defined template and collect its .Values usage.
func (s *Unittest) TestInclude_TraversesDefinedTemplate() {
	c := &chart.Chart{Templates: []*common.File{
		{
			Name: "templates/_defs.yaml",
			Data: []byte(`{{ define "tpl.a" }}{{ .Values.foo.bar }}{{ end }}`),
		},
		{
			Name: "templates/cm.yaml",
			Data: []byte(`
kind: ConfigMap
metadata: { name: x }
data:
  a: {{ include "tpl.a" . }}
`),
		},
	}}
	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)
	paths, err := parseFile("templates/cm.yaml", c.Templates[1].Data, idx)
	s.Require().NoError(err)

	s.Require().Len(paths, 1)
	s.Require().Equal("/foo/bar", paths[0].ID())
	s.Require().Equal("/K/K", paths[0].KindsString())
}

// Recursive includes should panic in Strict mode to prevent infinite loops.
func (s *Unittest) TestInclude_RecursionPanics() {
	c := &chart.Chart{Templates: []*common.File{
		{
			Name: "templates/_a.yaml",
			Data: []byte(`
{{ define "tpl.a" }}
{{ include "tpl.b" . }}
{{ end }}
`),
		},
		{
			Name: "templates/_b.yaml",
			Data: []byte(`
{{ define "tpl.b" }}
{{ include "tpl.a" . }}
{{ end }}
`),
		},
		{
			Name: "templates/cm.yaml",
			Data: []byte(`
kind: ConfigMap
metadata: { name: x }
data:
  a: {{ include "tpl.a" . }}
`),
		},
	}}
	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)
	s.Require().Panics(func() {
		_, _ = parseFile("templates/cm.yaml", c.Templates[2].Data, idx)
	})
}

// Excessive include depth should panic to guard against pathological charts.
func (s *Unittest) TestInclude_MaxDepthPanics() {
	// Build a chain tpl.0 -> tpl.1 -> ... -> tpl.65 to exceed includeMaxDepth=64

	var lines []string
	var defs []*common.File
	for i := 0; i < 66; i++ {
		name := fmt.Sprintf("tpl.%d", i)
		lines = append(lines, fmt.Sprintf(
			`{{ define "%s" }}{{ include "tpl.%d" . }}{{ end }}`,
			name,
			i+1,
		))
	}
	lines = append(lines, fmt.Sprintf(`{{ define "%s" }}x{{ end }}`, "tpl.end"))
	for i, line := range lines {
		defs = append(defs, &common.File{
			Name: fmt.Sprintf("templates/_%d.yaml", i),
			Data: []byte(line),
		})
	}
	main := &common.File{
		Name: "templates/cm.yaml",
		Data: []byte(`
kind: ConfigMap
metadata: { name: x }
data:
  a: {{ include "tpl.0" . }}
`),
	}
	c := &chart.Chart{Templates: append(defs, main)}
	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)
	s.Require().Panics(func() {
		_, _ = parseFile(main.Name, main.Data, idx)
	})
}

// Test that include with $ argument clears the prefix even when called
// from within a with block.
func (s *Unittest) TestInclude_RootContextClearsPrefix() {
	c := &chart.Chart{Templates: []*common.File{
		{
			Name: "templates/_helpers.yaml",
			// Use relative path .name which should be prefixed by current context
			Data: []byte(`{{ define "test.tpl" }}{{ .name }}{{ end }}`),
		},
		{
			Name: "templates/main.yaml",
			// Call from within `with .Values.ics` but pass $ (root context)
			Data: []byte(`{{ with .Values.ics }}{{ include "test.tpl" $ }}{{ end }}`),
		},
	}}

	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	paths, err := parseFile("templates/main.yaml", c.Templates[1].Data, idx)
	s.Require().NoError(err)

	// Should have no paths:
	// - with .Values.ics doesn't emit (only emits when body uses it)
	// - include passes $ (root context), clearing the prefix
	// - .name inside template with no prefix emits nothing
	expected := path.Paths{}
	path.EqualPaths(s, expected, paths)
}

// Test that include with . argument preserves the current prefix
func (s *Unittest) TestInclude_DotContextPreservesPrefix() {
	c := &chart.Chart{Templates: []*common.File{
		{
			Name: "templates/_helpers.yaml",
			// Use relative path .name which should be prefixed by current context
			Data: []byte(`{{ define "test.tpl" }}{{ .name }}{{ end }}`),
		},
		{
			Name: "templates/main.yaml",
			// Call from within `with .Values.config` and pass . (current context)
			Data: []byte(`{{ with .Values.config }}{{ include "test.tpl" . }}{{ end }}`),
		},
	}}

	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	paths, err := parseFile("templates/main.yaml", c.Templates[1].Data, idx)
	s.Require().NoError(err)

	// Should have /config and /config/name (. preserves the config prefix)
	expected := path.Paths{
		path.NewPath("config"),
		path.NewPath("config", "name"),
	}
	path.EqualPaths(s, expected, paths)
}

// Test that include with .Values.foo sets foo as the prefix
func (s *Unittest) TestInclude_ExplicitContextSetsPrefix() {
	c := &chart.Chart{Templates: []*common.File{
		{
			Name: "templates/_helpers.yaml",
			// Use relative path .name which should be prefixed by current context
			Data: []byte(`{{ define "test.tpl" }}{{ .name }}{{ end }}`),
		},
		{
			Name: "templates/main.yaml",
			// Call include with explicit .Values.database context
			Data: []byte(`{{ include "test.tpl" .Values.database }}`),
		},
	}}

	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)

	paths, err := parseFile("templates/main.yaml", c.Templates[1].Data, idx)
	s.Require().NoError(err)

	// Should have /database and /database/name
	expected := path.Paths{
		path.NewPath("database"),
		path.NewPath("database", "name"),
	}
	path.EqualPaths(s, expected, paths)
}
