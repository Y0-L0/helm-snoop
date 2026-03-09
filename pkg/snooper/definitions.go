package snooper

import (
	"fmt"
	"os"

	"helm.sh/helm/v4/pkg/chart/common"

	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

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
