package snooper

import "github.com/y0-l0/helm-snoop/pkg/vpath"

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

func (c *Chart) hasFindings() bool {
	return c.Result != nil && c.Result.hasFindings()
}

func (cs Charts) HasFindings() bool {
	for _, c := range cs {
		if c.hasFindings() {
			return true
		}
	}
	return false
}

func (cs Charts) unused() int {
	n := 0
	for _, c := range cs {
		if c.Result != nil {
			n += len(c.Result.Unused)
		}
	}
	return n
}

func (cs Charts) undefined() int {
	n := 0
	for _, c := range cs {
		if c.Result != nil {
			n += len(c.Result.Undefined)
		}
	}
	return n
}

func (cs Charts) scanned() int {
	n := 0
	for _, c := range cs {
		if !c.Skip {
			n++
		}
	}
	return n
}

func (cs Charts) skipped() int {
	n := 0
	for _, c := range cs {
		if c.Skip {
			n++
		}
	}
	return n
}
