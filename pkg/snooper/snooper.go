package snooper

import (
	"fmt"
	"os"

	"helm.sh/helm/v4/pkg/chart/common"
	loader "helm.sh/helm/v4/pkg/chart/v2/loader"

	"github.com/y0-l0/helm-snoop/pkg/tplparser"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// Chart holds resolved per-chart configuration for analysis.
type Chart struct {
	Path        string
	Skip        bool
	Ignore      vpath.Paths
	ValuesFiles []string
	ExtraValues map[string]any
}

type SnoopFunc func([]Chart) (Results, error)

func Snoop(charts []Chart) (Results, error) {
	var results Results
	for _, chart := range charts {
		result, err := snoopChart(chart)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func snoopChart(cs Chart) (*Result, error) {
	chart, err := loader.Load(cs.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read the helm chart.\nerror: %w", err)
	}

	used, err := tplparser.GetUsages(chart)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze the helm chart.\nerror: %w", err)
	}

	defined, err := loadDefinitions(chart.Raw, cs.ValuesFiles, cs.ExtraValues)
	if err != nil {
		return nil, err
	}

	result := &Result{ChartName: chart.Name()}
	result.Referenced, result.Unused, result.Undefined = vpath.MergeJoinLoose(defined, used)

	if len(cs.Ignore) > 0 {
		result = filterIgnoredWithMerge(result, cs.Ignore)
	}

	return result, nil
}

// findRawFile returns the raw bytes of a file from the chart's Raw slice, or nil if not found.
func findRawFile(raw []*common.File, name string) []byte {
	for _, f := range raw {
		if f.Name == name {
			return f.Data
		}
	}
	return nil
}

// loadDefinitions collects all defined value paths from the chart's values.yaml,
// additional values files, and inline extra values.
func loadDefinitions(
	raw []*common.File,
	extraFiles []string,
	extraValues map[string]any,
) (vpath.Paths, error) {
	defined, err := vpath.GetDefinitions(findRawFile(raw, "values.yaml"), "values.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml.\nerror: %w", err)
	}

	for _, f := range extraFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file %s.\nerror: %w", f, err)
		}
		extra, err := vpath.GetDefinitions(data, f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse values file %s.\nerror: %w", f, err)
		}
		defined = append(defined, extra...)
	}

	if len(extraValues) > 0 {
		defined = append(defined, vpath.GetDefinitionsFromMap(extraValues, "extraValues")...)
	}

	return defined, nil
}

// filterIgnoredWithMerge removes paths matching ignorePaths using MergeJoinLoose.
func filterIgnoredWithMerge(result *Result, ignorePaths vpath.Paths) *Result {
	if len(ignorePaths) == 0 {
		return result
	}

	_, _, keptUnused := vpath.MergeJoinLoose(ignorePaths, result.Unused)
	_, _, keptUndefined := vpath.MergeJoinLoose(ignorePaths, result.Undefined)

	return &Result{
		ChartName:  result.ChartName,
		Referenced: result.Referenced, // Never filtered
		Unused:     keptUnused,
		Undefined:  keptUndefined,
	}
}
