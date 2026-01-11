package snooper

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

type Result struct {
	Referenced     path.Paths
	DefinedNotUsed path.Paths
	UsedNotDefined path.Paths
}

type ResultsJSON struct {
	Referenced     path.PathsJSON `json:"referenced"`
	DefinedNotUsed path.PathsJSON `json:"definedNotUsed"`
	UsedNotDefined path.PathsJSON `json:"usedNotDefined"`
}

func (r *Result) HasFindings() bool {
	return len(r.DefinedNotUsed) > 0 || len(r.UsedNotDefined) > 0
}

func (r *Result) ToText(w io.Writer, showReferenced bool) error {
	if showReferenced {
		fmt.Fprintln(w, "Referenced:")
		for _, p := range r.Referenced {
			fmt.Fprintf(w, "  %s\n", p.ID())
		}
	}

	fmt.Fprintln(w, "Defined-not-used:")
	for _, p := range r.DefinedNotUsed {
		fmt.Fprintf(w, "  %s\n", p.ID())
	}

	fmt.Fprintln(w, "Used-not-defined:")
	for _, p := range r.UsedNotDefined {
		fmt.Fprintf(w, "  %s\n", p.ID())
	}

	return nil
}

func (r *Result) ToJSON(w io.Writer, showReferenced bool) error {
	resultsJSON := ResultsJSON{
		DefinedNotUsed: r.DefinedNotUsed.ToJSON(),
		UsedNotDefined: r.UsedNotDefined.ToJSON(),
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
		Referenced:     r.Referenced.ToJSON(),
		DefinedNotUsed: r.DefinedNotUsed.ToJSON(),
		UsedNotDefined: r.UsedNotDefined.ToJSON(),
	}
}
