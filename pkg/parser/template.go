package parser

import (
	"fmt"
	"text/template/parse"

	chart "helm.sh/helm/v4/pkg/chart/v2"
)

// TemplateDef captures a defined template's origin and parse tree root.
type TemplateDef struct {
	name string
	file string
	root *parse.ListNode
	tree *parse.Tree
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

// add inserts a template definition, panicking on duplicates.
func (ti *TemplateIndex) add(name string, def TemplateDef) {
	if _, exists := ti.byName[name]; exists {
		Must("duplicate template name: " + name)
	}
	ti.byName[name] = def
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
// prefix indicates the synthetic path prefix for dependency files (e.g., charts/<dep>/...).
func buildIndexRecursive(ch *chart.Chart, prefix string, idx *TemplateIndex, seen map[*chart.Chart]bool) error {
	if ch == nil {
		return nil
	}
	if seen[ch] {
		return nil
	}
	seen[ch] = true
	for _, tmpl := range ch.Templates {
		trees, err := parse.Parse(tmpl.Name, string(tmpl.Data), "", "", stubFuncMap)
		if err != nil {
			return err
		}
		for name, tree := range trees {
			if name == tmpl.Name {
				continue
			}
			if tree == nil || tree.Root == nil {
				continue
			}
			idx.add(name, TemplateDef{name: name, file: prefix + tmpl.Name, root: tree.Root, tree: tree})
		}
	}
	// Recurse into direct dependencies (Helm v4 exposes this as a method)
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
