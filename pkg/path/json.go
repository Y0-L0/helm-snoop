package path

type PathContextJSON struct {
	File     string `json:"file"`
	Template string `json:"template,omitempty"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

// PathJSON is a compact, stable JSON representation of a Path.
type PathJSON struct {
	ID       string            `json:"id"`
	Kinds    string            `json:"kinds"`
	Contexts []PathContextJSON `json:"contexts,omitempty"`
}

type PathsJSON []PathJSON

func (p Path) ToJSON() PathJSON {
	var contexts []PathContextJSON
	if len(p.Contexts) > 0 {
		contexts = make([]PathContextJSON, len(p.Contexts))
		for i, c := range p.Contexts {
			contexts[i] = c.ToJSON()
		}
	}
	return PathJSON{ID: p.ID(), Kinds: p.KindsString(), Contexts: contexts}
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
