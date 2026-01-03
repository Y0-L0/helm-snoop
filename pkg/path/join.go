package path

import (
	"slices"
	"sort"
)

// SortDedup returns a new slice sorted by Path.Compare and deduplicated.
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
	out = slices.CompactFunc(out, func(a, b *Path) bool { return a.Compare(*b) == 0 })
	return out
}

// CompareTokens compares two paths by tokens only, ignoring kinds.
func CompareTokens(a, b *Path) int { return slices.Compare(a.tokens, b.tokens) }

// equalKindLoose returns true if kinds are equal, or either side is anyKind.
func equalKindLoose(ka, kb kind) bool { return ka == kb || ka == anyKind || kb == anyKind }

// EqualLoose returns true if tokens are equal and per-segment kinds are equal
// or one side uses anyKind.
func EqualLoose(a, b *Path) bool {
	if CompareTokens(a, b) != 0 {
		return false
	}
	if len(a.kinds) != len(b.kinds) {
		return false
	}
	for i := range a.kinds {
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

	// For each path in a, check all of b and immediately classify as inter or onlyA
	for _, pa := range a {
		matched := false
		for j, pb := range b {
			if EqualLoose(pa, pb) {
				matchedB[j] = true
				matched = true
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

	// Results are already sorted (inputs were sorted by SortDedup)
	// but we deduplicate inter/onlyA/onlyB just in case
	return SortDedup(inter), SortDedup(onlyA), SortDedup(onlyB)
}
