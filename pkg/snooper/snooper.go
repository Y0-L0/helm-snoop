package snooper

import (
	"fmt"

	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/path"
	loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

type SnoopFunc func(string, path.Paths) (*Result, error)

func Snoop(chartPath string, ignorePaths path.Paths) (*Result, error) {
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the helm chart.\nerror: %w", err)
	}

	used, err := parser.GetUsages(chart)
	if err != nil {
		return nil, fmt.Errorf("Failed to analyze the helm chart.\nerror: %w", err)
	}

	defined := path.Paths{}
	if chart.Values != nil {
		path.GetDefinitions(path.Path{}, chart.Values, &defined)
	}

	result := &Result{}
	result.Referenced, result.Unused, result.Undefined = path.MergeJoinLoose(defined, used)

	if len(ignorePaths) > 0 {
		result = filterIgnoredWithMerge(result, ignorePaths)
	}

	return result, nil
}

// filterIgnoredWithMerge removes paths matching ignorePaths using MergeJoinLoose.
func filterIgnoredWithMerge(result *Result, ignorePaths path.Paths) *Result {
	if len(ignorePaths) == 0 {
		return result
	}

	_, _, keptUnused := path.MergeJoinLoose(ignorePaths, result.Unused)
	_, _, keptUndefined := path.MergeJoinLoose(ignorePaths, result.Undefined)

	return &Result{
		Referenced: result.Referenced, // Never filtered
		Unused:     keptUnused,
		Undefined:  keptUndefined,
	}
}
