package path

import (
	"slices"
	"sort"
)

// SortDedup sorts the slice in-place by Path.Compare.
// Removes duplicates and nil values.
func SortDedup(ps Paths) Paths {
	if len(ps) == 0 {
		return nil
	}
	// Copy and drop nils to avoid mutating the caller's backing array.
	out := make(Paths, 0, len(ps))
	for _, p := range ps {
		if p != nil {
			out = append(out, p)
		}
	}
	// Sort by value using the existing Less via sort.Sort.
	sort.Sort(out)
	// Deduplicate adjacent equals by value (tokens+kinds equality).
	out = slices.CompactFunc(out, func(a, b *Path) bool { return a.Compare(*b) == 0 })
	return out
}

// MergeJoinSet performs a set join between a and b using Path.Compare equality.
// Inputs are sorted and deduplicated in-place.
// Outputs (all sorted ascending):
//   - inter: elements present in both a and b
//   - onlyA: elements only in a
//   - onlyB: elements only in b
func MergeJoinSet(a, b Paths) (inter Paths, onlyA Paths, onlyB Paths) {
	// sort+dedup on copies to avoid mutating inputs
	a = SortDedup(a)
	b = SortDedup(b)

	inter = make(Paths, 0)
	onlyA = make(Paths, 0)
	onlyB = make(Paths, 0)

	i, j := 0, 0
	for i < len(a) && j < len(b) {
		comparison := a[i].Compare(*b[j])
		switch {
		case comparison == 0:
			inter = append(inter, a[i])
			i++
			j++
		case comparison < 0:
			onlyA = append(onlyA, a[i])
			i++
		default: // cmp > 0
			onlyB = append(onlyB, b[j])
			j++
		}
	}
	for ; i < len(a); i++ {
		onlyA = append(onlyA, a[i])
	}
	for ; j < len(b); j++ {
		onlyB = append(onlyB, b[j])
	}
	return
}
