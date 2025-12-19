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

// MergeJoinLoose performs an outer join grouped by tokens and matched loosely by kind.
func MergeJoinLoose(a, b Paths) (inter Paths, onlyA Paths, onlyB Paths) {
	a = SortDedup(a)
	b = SortDedup(b)

	inter, onlyA, onlyB = make(Paths, 0), make(Paths, 0), make(Paths, 0)
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch cmpTok := CompareTokens(a[i], b[j]); {
		case cmpTok < 0:
			onlyA = append(onlyA, a[i])
			i++
		case cmpTok > 0:
			onlyB = append(onlyB, b[j])
			j++
		default:
			ia0 := i
			for i < len(a) && CompareTokens(a[i], a[ia0]) == 0 {
				i++
			}
			jb0 := j
			for j < len(b) && CompareTokens(b[j], b[jb0]) == 0 {
				j++
			}
			matchedB := make([]bool, j-jb0)
			for ai := ia0; ai < i; ai++ {
				matched := false
				for bj := jb0; bj < j; bj++ {
					idx := bj - jb0
					if matchedB[idx] {
						continue
					}
					if EqualLoose(a[ai], b[bj]) {
						matchedB[idx] = true
						inter = append(inter, a[ai])
						matched = true
						break
					}
				}
				if !matched {
					onlyA = append(onlyA, a[ai])
				}
			}
			for bj := jb0; bj < j; bj++ {
				idx := bj - jb0
				if !matchedB[idx] {
					onlyB = append(onlyB, b[bj])
				}
			}
		}
	}
	for ; i < len(a); i++ {
		onlyA = append(onlyA, a[i])
	}
	for ; j < len(b); j++ {
		onlyB = append(onlyB, b[j])
	}

	sort.Sort(inter)
	sort.Sort(onlyA)
	sort.Sort(onlyB)
	return
}
