package vpath

import (
	"slices"
	"sort"
)

// subsumes returns true if a subsumes b (i.e., b is redundant if a exists).
// Specifically: /foo/* subsumes /foo, /foo/bar/* subsumes /foo/bar, etc.
func subsumes(a, b *Path) bool {
	aLen := len(a.kinds)
	bLen := len(b.kinds)

	// Check if a has terminal wildcard
	if aLen == 0 || a.kinds[aLen-1] != wildcardKind {
		return false
	}

	// a is /foo/*, b must be /foo (same prefix, no wildcard)
	if bLen != aLen-1 {
		return false
	}

	// Check all tokens match (excluding wildcard)
	for i := range bLen {
		if a.tokens[i] != b.tokens[i] || a.kinds[i] != b.kinds[i] {
			return false
		}
	}

	return true
}

// SortDedup returns a new slice sorted by Path.Compare and deduplicated.
// Merges contexts from exact duplicates and subsumed paths (e.g., /foo into /foo/*).
// Deduplicates contexts on each resulting path.
func SortDedup(ps Paths) Paths {
	if len(ps) == 0 {
		return nil
	}
	out := make(Paths, 0, len(ps))
	for _, p := range ps {
		if p != nil {
			out = append(out, p)
		}
	}
	sort.Sort(out)

	deduped := mergeDuplicates(out)
	filtered := removeSubsumed(deduped)

	for _, p := range filtered {
		p.Contexts = p.Contexts.Deduplicate()
	}
	return filtered
}

// mergeDuplicates collapses sorted paths with identical tokens+kinds.
// Merges contexts; Consumed wins over Checked.
func mergeDuplicates(sorted Paths) Paths {
	out := make(Paths, 0, len(sorted))
	for _, p := range sorted {
		if n := len(out); n > 0 && out[n-1].Compare(*p) == 0 {
			out[n-1].Contexts = append(out[n-1].Contexts, p.Contexts...)
			if p.Usage < out[n-1].Usage {
				out[n-1].Usage = p.Usage
			}
		} else {
			out = append(out, p)
		}
	}
	return out
}

// removeSubsumed drops paths subsumed by wildcards (e.g., /foo into /foo/*).
func removeSubsumed(ps Paths) Paths {
	out := make(Paths, 0, len(ps))
	for _, p := range ps {
		subsumed := false
		for _, other := range ps {
			if subsumes(other, p) {
				other.Contexts = append(other.Contexts, p.Contexts...)
				subsumed = true
				break
			}
		}
		if !subsumed {
			out = append(out, p)
		}
	}
	return out
}

// CompareTokens compares two paths by tokens only, ignoring kinds.
func CompareTokens(a, b *Path) int { return slices.Compare(a.tokens, b.tokens) }

// equalKindLoose returns true if kinds are equal, or either side is anyKind.
func equalKindLoose(ka, kb kind) bool { return ka == kb || ka == anyKind || kb == anyKind }

// getCompareLen calculates how many positions to compare between two paths.
// Returns (compareLen, ok) where ok=false means paths are incompatible.
func equalLenLoose(a, b *Path) (int, bool) {
	aLen := len(a.kinds)
	bLen := len(b.kinds)
	aHasTerminal := aLen > 0 && a.kinds[aLen-1] == wildcardKind
	bHasTerminal := bLen > 0 && b.kinds[bLen-1] == wildcardKind

	aEffective := aLen
	if aHasTerminal {
		aEffective--
	}
	bEffective := bLen
	if bHasTerminal {
		bEffective--
	}

	switch {
	case aHasTerminal && bHasTerminal:
		if aEffective < bEffective {
			return aEffective, true
		}
		return bEffective, true
	case aHasTerminal:
		if bEffective < aEffective {
			return 0, false
		}
		return aEffective, true
	case bHasTerminal:
		if aEffective < bEffective {
			return 0, false
		}
		return bEffective, true
	default:
		if aEffective != bEffective {
			return 0, false
		}
		return aEffective, true
	}
}

// EqualLoose returns true if paths match with:
// - Exact tokens and loose kind matching (anyKind matches anything), OR
// - Wildcard matching: terminal /* matches descendants, interior /* matches one segment.
func EqualLoose(a, b *Path) bool {
	compareLen, ok := equalLenLoose(a, b)
	if !ok {
		return false
	}

	for i := range compareLen {
		// If either position is a wildcard, automatic match
		if a.kinds[i] == wildcardKind || b.kinds[i] == wildcardKind {
			continue
		}
		if a.tokens[i] != b.tokens[i] {
			return false
		}
		if !equalKindLoose(a.kinds[i], b.kinds[i]) {
			return false
		}
	}

	return true
}

// MergeJoinLoose performs an outer join with loose matching by kind.
// Uses many-to-many matching: one path can match multiple paths (e.g., anyKind matches both indexKind and keyKind).
// Simple O(n*m) implementation is acceptable for typical helm chart path counts.
func MergeJoinLoose(a, b Paths) (inter Paths, onlyA Paths, onlyB Paths) {
	a = SortDedup(a)
	b = SortDedup(b)

	matchedB := make([]bool, len(b))

	for _, pa := range a {
		matched := false
		for j, pb := range b {
			if EqualLoose(pa, pb) {
				matchedB[j] = true
				matched = true
				pa.Contexts = append(pa.Contexts, pb.Contexts...)
			}
		}
		if matched {
			inter = append(inter, pa)
		} else {
			onlyA = append(onlyA, pa)
		}
	}

	// Collect unmatched paths from b
	for j, pb := range b {
		if !matchedB[j] {
			onlyB = append(onlyB, pb)
		}
	}

	return inter, onlyA, onlyB
}
