package snooper

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

type Result struct {
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

func (r *Result) ToText(w io.Writer, showReferenced bool) error {
	if showReferenced {
		fmt.Fprintln(w, "Referenced:")
		for _, p := range r.Referenced {
			printPathWithContext(w, p)
		}
	}

	fmt.Fprintln(w, "Unused:")
	for _, p := range r.Unused {
		printPathWithContext(w, p)
	}

	fmt.Fprintln(w, "Undefined:")
	for _, p := range r.Undefined {
		printPathWithContext(w, p)
	}

	return nil
}

func printPathWithContext(w io.Writer, p *path.Path) {
	fmt.Fprintf(w, "  %s\n", p.ID())
	for _, ctx := range p.Contexts {
		fmt.Fprintf(w, "    %s\n", ctx.String())
	}
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
