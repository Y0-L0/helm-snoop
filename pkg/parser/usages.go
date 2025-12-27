package parser

import (
	"fmt"
	"log/slog"
	"text/template/parse"

	"github.com/y0-l0/helm-snoop/pkg/path"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

// getUsages walks all chart templates and returns a flat list of observed .Values paths.
func GetUsages(ch *chart.Chart) (path.Paths, error) {
	result := make(path.Paths, 0)

	// Build template index for include resolution across files
	idx, err := BuildTemplateIndex(ch)
	if err != nil {
		return nil, err
	}
	slog.Debug("Built template index", "template_count", len(idx.byName))

	for _, tmpl := range ch.Templates {
		paths, err := parseFile(tmpl.Name, tmpl.Data, idx)
		slog.Debug("Analized template file", "name", tmpl.Name, "paths", paths)
		if err != nil {
			return nil, err
		}
		result = append(result, paths...)
	}
	return result, nil
}

// parseFile parses a file; idx enables include resolution across templates.
func parseFile(name string, data []byte, idx *TemplateIndex) (path.Paths, error) {
	trees, err := parse.Parse(name, string(data), "", "", stubFuncMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	slog.Debug("Parsed template file to a map of parse.Trees", "name", name)
	out := path.Paths{}
	for i, tree := range trees {
		slog.Debug("Analizing parse tree", "index", i, "root", tree.Root)
		ctx := newEvalCtx(tree, &out, idx)
		ctx.Eval(tree.Root)
	}
	return out, nil
}
