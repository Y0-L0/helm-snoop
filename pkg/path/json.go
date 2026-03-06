package path

type ContextJSON struct {
	File     string `json:"file"`
	Template string `json:"template,omitempty"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

// EntryJSON is a compact, stable JSON representation of a Path.
type EntryJSON struct {
	ID       string        `json:"id"`
	Kinds    string        `json:"kinds"`
	Contexts []ContextJSON `json:"contexts,omitempty"`
}

type EntriesJSON []EntryJSON

func (p Path) ToJSON() EntryJSON {
	var contexts []ContextJSON
	if len(p.Contexts) > 0 {
		contexts = make([]ContextJSON, len(p.Contexts))
		for i, c := range p.Contexts {
			contexts[i] = c.ToJSON()
		}
	}
	return EntryJSON{ID: p.ID(), Kinds: p.KindsString(), Contexts: contexts}
}

// ToJSON converts Paths into its compact JSON representation.
// Uses SortDedup to produce a sorted, deduplicated representation (non-mutating).
func (ps Paths) ToJSON() EntriesJSON {
	sorted := SortDedup(ps)

	out := make(EntriesJSON, 0, len(sorted))
	for _, p := range sorted {
		out = append(out, p.ToJSON())
	}
	return out
}
