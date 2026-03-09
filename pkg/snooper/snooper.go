// Package snooper analyzes Helm charts for unused and undefined value paths.
package snooper

import (
	"fmt"

	loader "helm.sh/helm/v4/pkg/chart/v2/loader"

	"github.com/y0-l0/helm-snoop/pkg/tplparser"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

type SnoopFunc func(Charts) error

func Snoop(charts Charts) error {
	for _, c := range charts {
		name, result, err := snoopChart(c)
		if err != nil {
			return err
		}
		c.Name = name
		c.Result = result
	}
	return nil
}

func snoopChart(cs *Chart) (string, *Result, error) {
	chart, err := loader.Load(cs.Path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read the helm chart.\nerror: %w", err)
	}

	if cs.Skip {
		return chart.Name(), nil, nil
	}

	used, err := tplparser.GetUsages(chart)
	if err != nil {
		return "", nil, fmt.Errorf("failed to analyze the helm chart.\nerror: %w", err)
	}

	defined, err := loadDefinitions(chart.Raw, cs.ValuesFiles, cs.ExtraValues)
	if err != nil {
		return "", nil, err
	}

	result := &Result{}
	result.Referenced, result.Unused, result.Undefined = vpath.MergeJoinLoose(defined, used)

	if len(cs.Ignore) > 0 {
		result = filterIgnoredWithMerge(result, cs.Ignore)
	}

	return chart.Name(), result, nil
}
