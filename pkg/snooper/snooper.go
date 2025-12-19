package snooper

import (
	"sort"

	parser "github.com/y0-l0/helm-snoop/pkg/oldparser"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

// Analyse analyses a Helm chart loaded via Helm's loader.
func Analyse(chart *chart.Chart) (*Result, error) {
	if chart == nil {
		panic("chart is nil")
	}

	// Per-file extraction into a flat list; reduce later for dedupe.
	usages, err := parser.GetUsages(chart)
	if err != nil {
		return nil, err
	}
	usedSet := reduceUsed(usages)

	// Flatten defaults from chart values
	definedSet := map[string]struct{}{}
	if chart.Values != nil {
		parser.GetDefinitions("", chart.Values, definedSet)
	}

	// Build result
	result := &Result{}
	result.Referenced = setToSortedSlice(usedSet)
	result.DefinedNotUsed = diffSets(definedSet, usedSet)
	result.UsedNotDefined = diffSets(usedSet, definedSet)
	sort.Strings(result.DefinedNotUsed)
	sort.Strings(result.UsedNotDefined)

	return result, nil
}

// reduceUsed converts a list of paths into a de-duplicated set.
func reduceUsed(paths []string) map[string]struct{} {
	set := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		set[p] = struct{}{}
	}
	return set
}

func setToSortedSlice(m map[string]struct{}) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func diffSets(a, b map[string]struct{}) []string {
	// return keys in a that are not in b
	if len(a) == 0 {
		return nil
	}
	out := make([]string, 0)
	for k := range a {
		if _, ok := b[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
