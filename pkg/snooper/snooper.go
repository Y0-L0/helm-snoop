package snooper

import (
	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/path"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

// Snoop analyses a Helm chart loaded via Helm's loader.
func Snoop(chart *chart.Chart) (*Result, error) {
	if chart == nil {
		panic("chart is nil")
	}

	used, err := parser.GetUsages(chart)
	if err != nil {
		return nil, err
	}

	defined := path.Paths{}
	if chart.Values != nil {
		path.GetDefinitions(path.Path{}, chart.Values, &defined)
	}

	result := &Result{}
	result.Referenced, result.DefinedNotUsed, result.UsedNotDefined = path.MergeJoinSet(defined, used)

	return result, nil
}
