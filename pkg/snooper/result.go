package snooper

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

type Result struct {
	Referenced vpath.Paths
	Unused     vpath.Paths
	Undefined  vpath.Paths
}

func (cs Charts) ToText(w io.Writer) {
	for _, c := range cs {
		formatChartCompact(w, c)
	}
	formatSummary(w, cs)
}

type ResultJSON struct {
	ChartName  string          `json:"chartName"`
	Referenced vpath.PathsJSON `json:"referenced,omitempty"`
	Unused     vpath.PathsJSON `json:"unused"`
	Undefined  vpath.PathsJSON `json:"undefined"`
}

func (r *Result) HasFindings() bool {
	return len(r.Unused) > 0 || len(r.Undefined) > 0
}

func (c *Chart) HasFindings() bool {
	return c.Result != nil && c.Result.HasFindings()
}

func (cs Charts) HasFindings() error {
	for _, c := range cs {
		if c.HasFindings() {
			return errors.New("")
		}
	}
	return nil
}

func (cs Charts) ToJSON(w io.Writer, showReferenced bool) error {
	var out []ResultJSON
	for _, c := range cs {
		rj := ResultJSON{
			ChartName: c.Name,
		}
		if c.Result != nil {
			rj.Unused = c.Result.Unused.ToJSON()
			rj.Undefined = c.Result.Undefined.ToJSON()
			if showReferenced {
				rj.Referenced = c.Result.Referenced.ToJSON()
			}
		}
		out = append(out, rj)
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}

// toJSON for backward compatibility with tests.
func (c *Chart) toJSON() ResultJSON {
	rj := ResultJSON{
		ChartName: c.Name,
	}
	if c.Result != nil {
		rj.Referenced = c.Result.Referenced.ToJSON()
		rj.Unused = c.Result.Unused.ToJSON()
		rj.Undefined = c.Result.Undefined.ToJSON()
	}
	return rj
}
