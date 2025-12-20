package path

// PathJSON is a compact, stable JSON representation of a Path.
type PathJSON struct {
	ID    string `json:"id"`
	Kinds string `json:"kinds"`
}

type PathsJSON []PathJSON

func (p Path) ToJSON() PathJSON {
	return PathJSON{ID: p.ID(), Kinds: p.KindsString()}
}

// ToJSON converts Paths into its compact JSON representation.
// Uses SortDedup to produce a sorted, deduplicated representation (non-mutating).
func (ps Paths) ToJSON() PathsJSON {
	sorted := SortDedup(ps)

	out := make(PathsJSON, 0, len(sorted))
	for _, p := range sorted {
		out = append(out, p.ToJSON())
	}
	return out
}
