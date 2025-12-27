package parser

import (
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

// Root includes template defined in immediate dependency.
func (s *Unittest) TestTemplateIndex_DependencyInclude() {
	s.T().Skip("Include is not yet implemented")
	root := &chart.Chart{Metadata: &chart.Metadata{Name: "root"}}
	child := &chart.Chart{Metadata: &chart.Metadata{Name: "child"}}

	child.Templates = []*common.File{
		{Name: "templates/_defs.tpl", Data: []byte(`{{ define "child.tpl.x" }}{{ .Values.child.k }}{{ end }}`)},
	}
	root.Templates = []*common.File{
		{Name: "templates/cm.yaml", Data: []byte("apiVersion: v1\nkind: ConfigMap\ndata:\n  v: {{ include \"child.tpl.x\" . }}\n")},
	}
	// Manually build index for root and child (simulate deps)
	idx := &TemplateIndex{byName: make(map[string]TemplateDef)}
	seen := make(map[*chart.Chart]bool)
	s.Require().NoError(buildIndexRecursive(root, "", idx, seen))
	s.Require().NoError(buildIndexRecursive(child, "charts/child/", idx, seen))

	paths, err := parseFile("templates/cm.yaml", root.Templates[0].Data, idx)
	s.Require().NoError(err)

	s.Require().Len(paths, 1)
	s.Equal("/child/k", paths[0].ID())
	s.Equal("/K/K", paths[0].KindsString())
}

// Root includes template defined in transitive dependency (child -> grandchild).
func (s *Unittest) TestTemplateIndex_TransitiveDependencyInclude() {
	s.T().Skip("Include is not yet implemented")
	root := &chart.Chart{Metadata: &chart.Metadata{Name: "root"}}
	child := &chart.Chart{Metadata: &chart.Metadata{Name: "child"}}
	grand := &chart.Chart{Metadata: &chart.Metadata{Name: "grand"}}

	grand.Templates = []*common.File{
		{Name: "templates/_defs.tpl", Data: []byte(`{{ define "grand.tpl.y" }}{{ .Values.grand.y }}{{ end }}`)},
	}
	child.Templates = []*common.File{{Name: "templates/other.yaml", Data: []byte("# no defines")}}
	root.Templates = []*common.File{
		{Name: "templates/cm.yaml", Data: []byte("apiVersion: v1\nkind: ConfigMap\ndata:\n  v: {{ include \"grand.tpl.y\" . }}\n")},
	}
	// Manually build index for root, child, and grand (simulate transitive deps)
	idx := &TemplateIndex{byName: make(map[string]TemplateDef)}
	seen := make(map[*chart.Chart]bool)
	s.Require().NoError(buildIndexRecursive(root, "", idx, seen))
	s.Require().NoError(buildIndexRecursive(child, "charts/child/", idx, seen))
	s.Require().NoError(buildIndexRecursive(grand, "charts/child/charts/grand/", idx, seen))
	paths, err := parseFile("templates/cm.yaml", root.Templates[0].Data, idx)
	s.Require().NoError(err)
	s.Require().Len(paths, 1)
	s.Equal("/grand/y", paths[0].ID())
	s.Equal("/K/K", paths[0].KindsString())
}

// Duplicate template names across dependencies should panic during index build.
func (s *Unittest) TestTemplateIndex_DuplicateNamesAcrossDeps() {
	root := &chart.Chart{Metadata: &chart.Metadata{Name: "root"}}
	d1 := &chart.Chart{Metadata: &chart.Metadata{Name: "d1"}}
	d2 := &chart.Chart{Metadata: &chart.Metadata{Name: "d2"}}

	d1.Templates = []*common.File{{Name: "templates/_defs.tpl", Data: []byte(`{{ define "dup.tpl" }}x{{ end }}`)}}
	d2.Templates = []*common.File{{Name: "templates/_defs.tpl", Data: []byte(`{{ define "dup.tpl" }}y{{ end }}`)}}

	idx := &TemplateIndex{byName: make(map[string]TemplateDef)}
	seen := make(map[*chart.Chart]bool)
	s.Require().Panics(func() {
		_ = buildIndexRecursive(root, "", idx, seen)
		_ = buildIndexRecursive(d1, "charts/d1/", idx, seen)
		_ = buildIndexRecursive(d2, "charts/d2/", idx, seen)
	})
}
