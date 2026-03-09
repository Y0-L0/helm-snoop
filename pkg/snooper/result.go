package snooper

import "github.com/y0-l0/helm-snoop/pkg/vpath"

type Result struct {
	Referenced vpath.Paths
	Unused     vpath.Paths
	Undefined  vpath.Paths
}

func (r *Result) hasFindings() bool {
	return len(r.Unused) > 0 || len(r.Undefined) > 0
}
