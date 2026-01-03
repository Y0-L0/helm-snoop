package path

func (s *Unittest) TestMergeJoinLoose_WildcardMatchesDirectChild() {
	a := Paths{np().Key("a").Wildcard()}
	b := Paths{np().Key("a").Key("b")}
	s.assertMergeJoinLoose(a, b, Paths{np().Key("a").Wildcard()}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_WildcardMatchesDeepDescendant() {
	a := Paths{np().Key("config").Wildcard()}
	b := Paths{np().Key("config").Key("nested").Key("value")}
	s.assertMergeJoinLoose(a, b, Paths{np().Key("config").Wildcard()}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_WildcardMatchesMultipleDescendants() {
	a := Paths{np().Key("a").Wildcard()}
	b := Paths{
		np().Key("a").Key("b"),
		np().Key("a").Key("c").Key("d"),
		np().Key("a").Key("x").Key("y").Key("z"),
	}
	s.assertMergeJoinLoose(a, b, Paths{np().Key("a").Wildcard()}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_WildcardDoesNotMatchDifferentPrefix() {
	a := Paths{np().Key("a").Wildcard()}
	b := Paths{np().Key("b").Key("c")}
	s.assertMergeJoinLoose(a, b, nil, Paths{np().Key("a").Wildcard()}, Paths{np().Key("b").Key("c")})
}

func (s *Unittest) TestMergeJoinLoose_WildcardDoesNotMatchShorterPath() {
	a := Paths{np().Key("a").Key("b").Wildcard()}
	b := Paths{np().Key("a")}
	s.assertMergeJoinLoose(a, b, nil, Paths{np().Key("a").Key("b").Wildcard()}, Paths{np().Key("a")})
}

func (s *Unittest) TestMergeJoinLoose_MultipleWildcardsMatchDifferentDescendants() {
	a := Paths{
		np().Key("a").Wildcard(),
		np().Key("b").Wildcard(),
	}
	b := Paths{
		np().Key("a").Key("x"),
		np().Key("b").Key("y").Key("z"),
		np().Key("c").Key("w"),
	}
	expInter := Paths{
		np().Key("a").Wildcard(),
		np().Key("b").Wildcard(),
	}
	s.assertMergeJoinLoose(a, b, expInter, nil, Paths{np().Key("c").Key("w")})
}

// Interior wildcard tests - must match exactly one segment
func (s *Unittest) TestMergeJoinLoose_InteriorWildcardMatchesOneSegment() {
	a := Paths{np().Key("a").Wildcard().Key("c")}
	b := Paths{np().Key("a").Key("b").Key("c")}
	s.assertMergeJoinLoose(a, b, Paths{a[0]}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_InteriorWildcardDoesNotMatchMultipleSegments() {
	a := Paths{np().Key("a").Wildcard().Key("c")}
	b := Paths{np().Key("a").Key("b").Key("d").Key("c")}
	s.assertMergeJoinLoose(a, b, nil, a, b)
}

// Multiple wildcards
func (s *Unittest) TestMergeJoinLoose_InteriorAndTerminalWildcard() {
	a := Paths{np().Key("a").Wildcard().Key("c").Wildcard()}
	b := Paths{np().Key("a").Key("b").Key("c").Key("d")}
	s.assertMergeJoinLoose(a, b, Paths{a[0]}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_InteriorAndTerminalWildcardDeep() {
	a := Paths{np().Key("a").Wildcard().Key("c").Wildcard()}
	b := Paths{np().Key("a").Key("b").Key("c").Key("d").Key("e")}
	s.assertMergeJoinLoose(a, b, Paths{a[0]}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_TwoInteriorWildcards() {
	a := Paths{np().Key("a").Wildcard().Wildcard().Key("d")}
	b := Paths{np().Key("a").Key("b").Key("c").Key("d")}
	s.assertMergeJoinLoose(a, b, Paths{a[0]}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_TwoInteriorWildcardsWrongLength() {
	a := Paths{np().Key("a").Wildcard().Wildcard().Key("d")}
	b := Paths{np().Key("a").Key("b").Key("d")}
	s.assertMergeJoinLoose(a, b, nil, a, b)
}

// Edge cases
func (s *Unittest) TestMergeJoinLoose_RootTerminalWildcard() {
	a := Paths{np().Wildcard()}
	b := Paths{np().Key("a"), np().Key("a").Key("b")}
	s.assertMergeJoinLoose(a, b, Paths{a[0]}, nil, nil)
}

func (s *Unittest) TestMergeJoinLoose_InteriorWildcardAtStart() {
	a := Paths{np().Wildcard().Key("b")}
	b := Paths{np().Key("a").Key("b")}
	s.assertMergeJoinLoose(a, b, Paths{a[0]}, nil, nil)
}
