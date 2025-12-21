package parser

import (
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
		must("duplicate template name: " + name)
	}
	ti.byName[name] = def
}

// empty reports whether the index is empty.
func (ti *TemplateIndex) empty() bool { return len(ti.byName) == 0 }

// BuildTemplateIndex parses all templates in the chart and indexes define'd templates by name.
func BuildTemplateIndex(ch *chart.Chart) (*TemplateIndex, error) {
	idx := &TemplateIndex{byName: make(map[string]TemplateDef)}
	for _, tmpl := range ch.Templates {
		trees, err := parse.Parse(tmpl.Name, string(tmpl.Data), "", "", templFuncMap)
		if err != nil {
			return nil, err
		}
		for name, tree := range trees {
			// Exclude the synthetic entry that matches the file name; we want only define'd names.
			if name == tmpl.Name {
				continue
			}
			if tree == nil || tree.Root == nil {
				continue
			}
			idx.add(name, TemplateDef{name: name, file: tmpl.Name, root: tree.Root, tree: tree})
		}
	}
	return idx, nil
}

// (no test-only helpers here; parser tests use real chart.Chart construction)
