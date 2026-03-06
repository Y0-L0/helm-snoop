package tplparser

import (
	"fmt"
	"log/slog"
	"text/template/parse"

	chart "helm.sh/helm/v4/pkg/chart/v2"

	"github.com/y0-l0/helm-snoop/internal/assert"
)

// TemplateDef captures a defined template's origin and parse tree root.
type TemplateDef struct {
	name      string
	file      string
	chartName string // name of the chart that defines this template
	prefix    string // dependency path prefix (e.g., "charts/mariadb/charts/common/")
	root      *parse.ListNode
	tree      *parse.Tree
}

// TemplateIndex provides lookup of defined templates by name across a chart.
type TemplateIndex struct {
	byName map[string]TemplateDef
}

// get returns the template definition by name, if present.
func (ti *TemplateIndex) get(name string) (TemplateDef, bool) {
	if ti == nil {
		return TemplateDef{}, false
	}
	d, ok := ti.byName[name]
	return d, ok
}

// add inserts a template definition; see helm-dependency-nightmare.md.
func (ti *TemplateIndex) add(name string, def TemplateDef) {
	previous, exists := ti.byName[name]
	ti.byName[name] = def
	if !exists {
		return
	}
	if previous.chartName != "" && previous.chartName == def.chartName && previous.prefix != def.prefix {
		slog.Info("duplicate template from shared dependency (last definition wins)",
			"name", name, "chart", def.chartName,
			"kept", def.file, "overwritten", previous.file)
		return
	}
	slog.Warn("duplicate template name", "name", name,
		"first", previous.file, "second", def.file)
	assert.Must("duplicate template name: " + name)
}

// addFromFile indexes all define'd templates from a parsed template file.
func (ti *TemplateIndex) addFromFile(fileName, chartName, prefix string, trees map[string]*parse.Tree) {
	for name, tree := range trees {
		if name == fileName || tree == nil || tree.Root == nil {
			continue
		}
		ti.add(name, TemplateDef{
			name:      name,
			file:      prefix + fileName,
			chartName: chartName,
			prefix:    prefix,
			root:      tree.Root,
			tree:      tree,
		})
	}
}

// indexChart parses and indexes all define'd templates in a single chart.
func (ti *TemplateIndex) indexChart(ch *chart.Chart, prefix string) error {
	chartName := ""
	if ch.Metadata != nil {
		chartName = ch.Metadata.Name
	}
	for _, tmpl := range ch.Templates {
		trees, err := parse.Parse(tmpl.Name, string(tmpl.Data), "", "", stubFuncMap)
		if err != nil {
			return err
		}
		ti.addFromFile(tmpl.Name, chartName, prefix, trees)
	}
	return nil
}

// empty reports whether the index is empty.
func (ti *TemplateIndex) empty() bool { return len(ti.byName) == 0 }

// BuildTemplateIndex parses all templates in the chart and indexes define'd templates by name.
func BuildTemplateIndex(ch *chart.Chart) (*TemplateIndex, error) {
	idx := &TemplateIndex{byName: make(map[string]TemplateDef)}
	seen := make(map[*chart.Chart]bool)
	if err := buildIndexRecursive(ch, "", idx, seen); err != nil {
		return nil, err
	}
	return idx, nil
}

// buildIndexRecursive adds define'd templates from chart and its transitive dependencies.
func buildIndexRecursive(ch *chart.Chart, prefix string, idx *TemplateIndex, seen map[*chart.Chart]bool) error {
	if ch == nil || seen[ch] {
		return nil
	}
	seen[ch] = true
	if err := idx.indexChart(ch, prefix); err != nil {
		return err
	}
	for _, dep := range ch.Dependencies() {
		depName := "unknown"
		if dep != nil && dep.Metadata != nil && dep.Metadata.Name != "" {
			depName = dep.Metadata.Name
		}
		depPrefix := fmt.Sprintf("%scharts/%s/", prefix, depName)
		if err := buildIndexRecursive(dep, depPrefix, idx, seen); err != nil {
			return err
		}
	}
	return nil
}

// (no test-only helpers here; parser tests use real chart.Chart construction)
