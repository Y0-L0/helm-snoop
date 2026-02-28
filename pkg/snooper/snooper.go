package snooper

import (
	"fmt"
	"os"

	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/path"
	"helm.sh/helm/v4/pkg/chart/common"
	loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

type SnoopFunc func(string, path.Paths, []string) (*Result, error)

func Snoop(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*Result, error) {
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the helm chart.\nerror: %w", err)
	}

	used, err := parser.GetUsages(chart)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze the helm chart.\nerror: %w", err)
	}

	defined, err := loadDefinitions(chart.Raw, valuesFiles)
	if err != nil {
		return nil, err
	}

	result := &Result{ChartName: chart.Name()}
	result.Referenced, result.Unused, result.Undefined = path.MergeJoinLoose(defined, used)

	if len(ignorePaths) > 0 {
		result = filterIgnoredWithMerge(result, ignorePaths)
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

// loadDefinitions collects all defined value paths from the chart's values.yaml
// and any additional values files provided on the command line.
func loadDefinitions(raw []*common.File, extraFiles []string) (path.Paths, error) {
	defined, err := path.GetDefinitions(findRawFile(raw, "values.yaml"), "values.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml.\nerror: %w", err)
	}

	for _, f := range extraFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read values file %s.\nerror: %w", f, err)
		}
		extra, err := path.GetDefinitions(data, f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse values file %s.\nerror: %w", f, err)
		}
		defined = append(defined, extra...)
	}

	return defined, nil
}

// filterIgnoredWithMerge removes paths matching ignorePaths using MergeJoinLoose.
func filterIgnoredWithMerge(result *Result, ignorePaths path.Paths) *Result {
	if len(ignorePaths) == 0 {
		return result
	}

	_, _, keptUnused := path.MergeJoinLoose(ignorePaths, result.Unused)
	_, _, keptUndefined := path.MergeJoinLoose(ignorePaths, result.Undefined)

	return &Result{
		ChartName:  result.ChartName,
		Referenced: result.Referenced, // Never filtered
		Unused:     keptUnused,
		Undefined:  keptUndefined,
	}
}
