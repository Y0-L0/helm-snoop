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

func (s *Unittest) TestSortDedup_MergesContexts() {
	ctxA := PathContext{FileName: "templates/deployment.yaml", Line: 10, Column: 5}
	ctxB := PathContext{FileName: "templates/service.yaml", Line: 20, Column: 3}

	p1 := NewPath("a", "b")
	p1.Contexts = Contexts{ctxA}
	p2 := NewPath("a", "b") // duplicate
	p2.Contexts = Contexts{ctxB}

	got := SortDedup(Paths{p1, p2})

	s.Require().Len(got, 1)
	s.Equal(".a.b", got[0].ID())
	s.Equal(Contexts{ctxA, ctxB}, got[0].Contexts)
}

func (s *Unittest) TestSortDedup_WildcardSubsumption_MergesContexts() {
	ctxFoo := PathContext{FileName: "templates/deployment.yaml", Line: 5, Column: 1}
	ctxFooWild := PathContext{FileName: "templates/service.yaml", Line: 10, Column: 1}

	pFoo := NewPath("foo")
	pFoo.Contexts = Contexts{ctxFoo}
	pFooWild := np().Key("foo").Wildcard()
	pFooWild.Contexts = Contexts{ctxFooWild}

	got := SortDedup(Paths{pFoo, pFooWild})

	// /foo should be subsumed by /foo/*, but its context should be merged
	s.Require().Len(got, 1)
	s.Equal(".foo.*", got[0].ID())
	s.Equal(Contexts{ctxFooWild, ctxFoo}, got[0].Contexts)
}

func (s *Unittest) TestSortDedup_DeduplicatesContexts() {
	ctx := PathContext{FileName: "values.yaml", Line: 5, Column: 3}

	p1 := NewPath("a")
	p1.Contexts = Contexts{ctx}
	p2 := NewPath("a") // duplicate with same context
	p2.Contexts = Contexts{ctx}

	got := SortDedup(Paths{p1, p2})

	s.Require().Len(got, 1)
	// Same context should be deduplicated
	s.Equal(Contexts{ctx}, got[0].Contexts)
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
	s.Equal(".items.0", inter[0].ID())

	// Nothing unused
	s.Equal(0, len(onlyDef), "no definitions should be unmatched")

	// Both usages matched the definition, so nothing undefined
	s.Equal(0, len(onlyUsage), "both usages should match the definition")
}
