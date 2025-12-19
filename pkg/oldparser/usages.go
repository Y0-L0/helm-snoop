package oldparser

import (
	"fmt"
	"text/template/parse"

	chart "helm.sh/helm/v4/pkg/chart/v2"
)

// getUsages walks all chart templates and returns a flat list of observed .Values paths.
func GetUsages(ch *chart.Chart) ([]string, error) {
	all := make([]string, 0)
	for _, tmpl := range ch.Templates {
		vals, err := parseFile(tmpl.Name, tmpl.Data)
		if err != nil {
			return nil, err
		}
		all = append(all, vals...)
	}
	return all, nil
}

// parseFile parses one template file and returns all observed .Values paths.
func parseFile(name string, data []byte) ([]string, error) {
	trees, err := parse.Parse(name, string(data), "", "", templFuncMap)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}
	out := make([]string, 0)
	for _, t := range trees {
		out = append(out, collectUsedValues(t.Root)...)
	}
	return out, nil
}
