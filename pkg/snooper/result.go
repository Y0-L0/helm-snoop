package snooper

import (
	"encoding/json"
	"io"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

type Result struct {
	ChartName  string
	Referenced path.Paths
	Unused     path.Paths
	Undefined  path.Paths
}

type ResultsJSON struct {
	Referenced path.PathsJSON `json:"referenced"`
	Unused     path.PathsJSON `json:"unused"`
	Undefined  path.PathsJSON `json:"undefined"`
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

// toJSON for backward compatibility with tests
func (r *Result) toJSON() ResultsJSON {
	return ResultsJSON{
		Referenced: r.Referenced.ToJSON(),
		Unused:     r.Unused.ToJSON(),
		Undefined:  r.Undefined.ToJSON(),
	}
}
