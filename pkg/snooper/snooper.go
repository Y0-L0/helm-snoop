package snooper

import (
	"fmt"
	"os"

	"helm.sh/helm/v4/pkg/chart/common"
	loader "helm.sh/helm/v4/pkg/chart/v2/loader"

	"github.com/y0-l0/helm-snoop/pkg/tplparser"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

// Chart holds resolved per-chart configuration and analysis results.
type Chart struct {
	Path        string
	Name        string
	Skip        bool
	Ignore      vpath.Paths
	ValuesFiles []string
	ExtraValues map[string]any
	Result      *Result
}

type Charts []*Chart

type SnoopFunc func(Charts) error

func Snoop(charts Charts) error {
	for _, c := range charts {
		if err := snoopChart(c); err != nil {
			return err
		}
	}
	return nil
}

func snoopChart(cs *Chart) error {
	chart, err := loader.Load(cs.Path)
	if err != nil {
		return fmt.Errorf("failed to read the helm chart.\nerror: %w", err)
	}

	cs.Name = chart.Name()

	if cs.Skip {
		return nil
	}

	used, err := tplparser.GetUsages(chart)
	if err != nil {
		return fmt.Errorf("failed to analyze the helm chart.\nerror: %w", err)
	}

	defined, err := loadDefinitions(chart.Raw, cs.ValuesFiles, cs.ExtraValues)
	if err != nil {
		return err
	}

	result := &Result{}
	result.Referenced, result.Unused, result.Undefined = vpath.MergeJoinLoose(defined, used)

	if len(cs.Ignore) > 0 {
		result = filterIgnoredWithMerge(result, cs.Ignore)
	}

	cs.Result = result
	return nil
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
		Referenced: result.Referenced, // Never filtered
		Unused:     keptUnused,
		Undefined:  keptUndefined,
	}
}
