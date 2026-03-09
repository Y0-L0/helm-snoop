package snooper

import "github.com/y0-l0/helm-snoop/pkg/vpath"

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
