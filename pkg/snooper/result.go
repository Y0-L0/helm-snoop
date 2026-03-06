package snooper

import (
	"encoding/json"
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

type ResultsJSON struct {
	Referenced vpath.PathsJSON `json:"referenced"`
	Unused     vpath.PathsJSON `json:"unused"`
	Undefined  vpath.PathsJSON `json:"undefined"`
}

func (r *Result) HasFindings() bool {
	return len(r.Unused) > 0 || len(r.Undefined) > 0
}

// ToText writes the result in compact text format.
func (r *Result) ToText(w io.Writer) {
	formatChartCompact(w, r)
}

func (r *Result) ToJSON(w io.Writer, showReferenced bool) error {
	resultsJSON := ResultsJSON{
		Unused:    r.Unused.ToJSON(),
		Undefined: r.Undefined.ToJSON(),
	}
	if showReferenced {
		resultsJSON.Referenced = r.Referenced.ToJSON()
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resultsJSON)
}

// toJSON for backward compatibility with tests.
func (r *Result) toJSON() ResultsJSON {
	return ResultsJSON{
		Referenced: r.Referenced.ToJSON(),
		Unused:     r.Unused.ToJSON(),
		Undefined:  r.Undefined.ToJSON(),
	}
}
