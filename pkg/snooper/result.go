package snooper

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

type Result struct {
	ChartName  string
	Referenced vpath.Paths
	Unused     vpath.Paths
	Undefined  vpath.Paths
}

type Results []*Result

func (rs Results) ToText(w io.Writer) {
	for _, r := range rs {
		formatChartCompact(w, r)
	}
	formatSummary(w, rs)
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

func (rs Results) HasFindings() error {
	for _, r := range rs {
		if r.HasFindings() {
			return errors.New("")
		}
	}
	return nil
}

func (rs Results) ToJSON(w io.Writer, showReferenced bool) error {
	var out []ResultJSON
	for _, r := range rs {
		rj := ResultJSON{
			ChartName: r.ChartName,
			Unused:    r.Unused.ToJSON(),
			Undefined: r.Undefined.ToJSON(),
		}
		if showReferenced {
			rj.Referenced = r.Referenced.ToJSON()
		}
		out = append(out, rj)
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(out)
}

// toJSON for backward compatibility with tests.
func (r *Result) toJSON() ResultJSON {
	return ResultJSON{
		ChartName:  r.ChartName,
		Referenced: r.Referenced.ToJSON(),
		Unused:     r.Unused.ToJSON(),
		Undefined:  r.Undefined.ToJSON(),
	}
}
