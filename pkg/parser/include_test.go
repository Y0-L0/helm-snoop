package parser

import (
	"fmt"
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"
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
