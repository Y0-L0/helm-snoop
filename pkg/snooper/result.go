package snooper

import "github.com/y0-l0/helm-snoop/pkg/path"

// Result holds the outcome of an analysis run.
// It is intended to be serialized or formatted by multiple output backends.
type Result struct {
	Referenced     path.Paths
	DefinedNotUsed path.Paths
	UsedNotDefined path.Paths
}
