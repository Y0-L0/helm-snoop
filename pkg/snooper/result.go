package snooper

import "github.com/y0-l0/helm-snoop/pkg/path"

// Result holds the outcome of an analysis run.
// It is intended to be serialized or formatted by multiple output backends.
type Result struct {
	Referenced     path.Paths
	DefinedNotUsed path.Paths
	UsedNotDefined path.Paths
}

// ResultsJSON is a compact JSON representation of a Result, exposing only stable path encodings.
type ResultsJSON struct {
	Referenced     path.PathsJSON `json:"referenced"`
	DefinedNotUsed path.PathsJSON `json:"definedNotUsed"`
	UsedNotDefined path.PathsJSON `json:"usedNotDefined"`
}

// ToJSON converts a Result into its compact, deterministic JSON representation.
func (r *Result) ToJSON() ResultsJSON {
	return ResultsJSON{
		Referenced:     r.Referenced.ToJSON(),
		DefinedNotUsed: r.DefinedNotUsed.ToJSON(),
		UsedNotDefined: r.UsedNotDefined.ToJSON(),
	}
}
