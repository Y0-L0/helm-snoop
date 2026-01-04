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

func (s *Unittest) TestSortDedup_WildcardSubsumption() {
	ps := Paths{
		NewPath("foo"),
		np().Key("foo").Wildcard(),
		NewPath("bar", "baz"),
		np().Key("bar").Key("baz").Wildcard(),
		NewPath("other"),
	}

	got := SortDedup(ps)

	// /foo should be removed (subsumed by /foo/*)
	// /bar/baz should be removed (subsumed by /bar/baz/*)
	// /other should remain (no wildcard version)
	expected := Paths{
		np().Key("bar").Key("baz").Wildcard(),
		np().Key("foo").Wildcard(),
		NewPath("other"),
	}
	EqualInorderPaths(s, expected, got)
}

func (s *Unittest) assertMergeJoin(a, b Paths, expInter, expOnlyA, expOnlyB Paths) {
	// Copies to assert non-mutation of inputs
	origA := append(Paths(nil), a...)
	origB := append(Paths(nil), b...)

	inter, onlyA, onlyB := MergeJoinLoose(a, b)

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

// Loose join tests with anyKind
func (s *Unittest) assertMergeJoinLoose(a, b Paths, expInter, expOnlyA, expOnlyB Paths) {
	origA := append(Paths(nil), a...)
	origB := append(Paths(nil), b...)
	inter, onlyA, onlyB := MergeJoinLoose(a, b)
	EqualInorderPaths(s, expInter, inter)
	EqualInorderPaths(s, expOnlyA, onlyA)
	EqualInorderPaths(s, expOnlyB, onlyB)
	s.Require().True(slices.Equal(a, origA))
	s.Require().True(slices.Equal(b, origB))
}

func (s *Unittest) TestMergeJoinLoose_AnyMatchesKeyAndIdx() {
	a := Paths{np().Any("x")}
	b := Paths{np().Key("x"), np().Idx("x")}
	// Many-to-many matching: anyKind matches BOTH keyKind and indexKind
	// So the one path in 'a' matches both paths in 'b'
	s.assertMergeJoinLoose(a, b, Paths{a[0]}, nil, nil)
	// Flipped: both paths in 'a' match the anyKind in 'b'
	// Results are sorted by kind: indexKind ('I') before keyKind ('K')
	s.assertMergeJoinLoose(b, a, Paths{np().Idx("x"), np().Key("x")}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_DisjointKinds() {
	a := Paths{np().Key("x")}
	b := Paths{np().Idx("x")}
	s.assertMergeJoinLoose(a, b, nil, a, b)
	s.assertMergeJoinLoose(b, a, nil, b, a)
}

// TestMergeJoinLoose_OneDefinitionMatchesMultipleUsages tests that a single
// definition can match multiple usages with different kinds.
//
// Scenario:
//
//	values.yaml defines: items: [foo, bar]  → creates /items/0 (indexKind)
//
//	templates use:
//	  {{ index .Values.items 0 }}        → /items/0 (indexKind)
//	  {{ index .Values.items .dynamic }} → /items/* (anyKind) - unknown index
//
// Expected: The definition /items/0 should match BOTH usages.
func (s *Unittest) TestMergeJoinLoose_OneDefinitionMatchesMultipleUsages() {
	definitions := Paths{
		np().Key("items").Idx("0"), // Defined in values.yaml
	}

	usages := Paths{
		np().Key("items").Idx("0"), // Used with literal index
		np().Key("items").Any("0"), // Used with dynamic index (anyKind)
	}

	inter, onlyDef, onlyUsage := MergeJoinLoose(definitions, usages)

	// The definition matches both usages, so it appears in inter
	s.Equal(1, len(inter), "definition should match both usages")
	s.Equal("/items/0", inter[0].ID())

	// Nothing defined but not used
	s.Equal(0, len(onlyDef), "no definitions should be unmatched")

	// Both usages matched the definition, so nothing used but not defined
	s.Equal(0, len(onlyUsage), "both usages should match the definition")
}
