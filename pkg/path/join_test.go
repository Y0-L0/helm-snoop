package path

import (
	"slices"
)

func (s *Unittest) TestSortDedup() {
	ps := Paths{
		NewPath("a", "b"),
		NewPath("a", "a"),
		NewPath("a", "b"), // duplicate
		NewPath("c"),
		nil, // should be skipped
	}
	// Keep a copy to verify non-mutation of input
	orig := append(Paths(nil), ps...)

	got := SortDedup(ps)

	EqualInorderPaths(s, Paths{NewPath("a", "a"), NewPath("a", "b"), NewPath("c")}, got)
	// Verify input slice not mutated (pointer identity and order)
	s.Require().True(slices.Equal(ps, orig), "input mutated")
}

func (s *Unittest) assertMergeJoin(a, b Paths, expInter, expOnlyA, expOnlyB Paths) {
	// Copies to assert non-mutation of inputs
	origA := append(Paths(nil), a...)
	origB := append(Paths(nil), b...)

	inter, onlyA, onlyB := MergeJoinSet(a, b)

	EqualInorderPaths(s, expInter, inter)
	EqualInorderPaths(s, expOnlyA, onlyA)
	EqualInorderPaths(s, expOnlyB, onlyB)

	// Ensure inputs are not mutated (pointer identity and order)
	s.Require().True(slices.Equal(a, origA), "input 'a' mutated")
	s.Require().True(slices.Equal(b, origB), "input 'b' mutated")
}

func (s *Unittest) TestMergeJoinSet() {
	a := Paths{
		NewPath("a"),
		NewPath("b"),
		NewPath("b"), // duplicate in a
		NewPath("d"),
	}
	b := Paths{
		NewPath("b"),
		NewPath("c"),
		NewPath("d"),
		NewPath("d"), // duplicate in b
	}

	expInter := Paths{NewPath("b"), NewPath("d")}
	expOnlyA, expOnlyB := Paths{NewPath("a")}, Paths{NewPath("c")}

	s.assertMergeJoin(a, b, expInter, expOnlyA, expOnlyB)
	s.assertMergeJoin(b, a, expInter, expOnlyB, expOnlyA)
}

func (s *Unittest) TestMergeJoinSet_DisjointKinds() {
	a := Paths{np().Key("1")}
	b := Paths{np().Idx("1")}

	expInter, expOnlyA, expOnlyB := Paths(nil), Paths{a[0]}, Paths{b[0]}

	s.assertMergeJoin(a, b, expInter, expOnlyA, expOnlyB)
	s.assertMergeJoin(b, a, expInter, expOnlyB, expOnlyA)
}

func (s *Unittest) TestMergeJoinSet_EmptySides() {
	a := Paths{}
	b := Paths{NewPath("x"), NewPath("x")}

	expInter, expOnlyA, expOnlyB := Paths(nil), Paths(nil), Paths{NewPath("x")}

	s.assertMergeJoin(a, b, expInter, expOnlyA, expOnlyB)
	s.assertMergeJoin(b, a, expInter, expOnlyB, expOnlyA)
}
