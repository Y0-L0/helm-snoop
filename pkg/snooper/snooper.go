package snooper

import (
	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/path"
	loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

type SnoopFunc func(string, []string) (*Result, error)

func Snoop(chartPath string, ignore []string) (*Result, error) {
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
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
	result.Referenced, result.DefinedNotUsed, result.UsedNotDefined = path.MergeJoinLoose(defined, used)

	if len(ignore) > 0 {
		result = filterIgnored(result, ignore)
	}

	return result, nil
}

// filterIgnored removes paths from DefinedNotUsed and UsedNotDefined only
func filterIgnored(result *Result, ignore []string) *Result {
	if len(ignore) == 0 {
		return result
	}

	ignoreMap := make(map[string]bool, len(ignore))
	for _, key := range ignore {
		ignoreMap[key] = true
	}

	filteredDefinedNotUsed := make(path.Paths, 0, len(result.DefinedNotUsed))
	for _, p := range result.DefinedNotUsed {
		if !ignoreMap[p.ID()] {
			filteredDefinedNotUsed = append(filteredDefinedNotUsed, p)
		}
	}

	filteredUsedNotDefined := make(path.Paths, 0, len(result.UsedNotDefined))
	for _, p := range result.UsedNotDefined {
		if !ignoreMap[p.ID()] {
			filteredUsedNotDefined = append(filteredUsedNotDefined, p)
		}
	}

	return &Result{
		Referenced:     result.Referenced,
		DefinedNotUsed: filteredDefinedNotUsed,
		UsedNotDefined: filteredUsedNotDefined,
	}
}
