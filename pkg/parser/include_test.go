package parser

import (
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
