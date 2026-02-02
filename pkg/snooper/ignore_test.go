package snooper

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
)

func np() *path.Path { return &path.Path{} }

func (s *Unittest) TestFilterIgnoredWithMerge_Unused() {
	unused := path.Paths{
		np().Key("image").Key("tag"),
		np().Key("config").Key("nested").Key("value"),
		np().Key("config").Key("other"),
		np().Key("replicas"),
		np().Key("items").Idx("0"),
		np().Key("items").Key("0"),
	}

	ignorePatterns := path.Paths{
		np().Key("image").Key("tag"),  // Exact match
		np().Key("config").Wildcard(), // Terminal wildcard
		np().Key("items").Any("0"),    // AnyKind
	}

	result := &Result{
		Referenced: path.Paths{np().Key("ref")},
		Unused:     unused,
		Undefined:  path.Paths{},
	}

	filtered := filterIgnoredWithMerge(result, ignorePatterns)

	s.Require().Equal(result.Referenced, filtered.Referenced)
	s.Require().Equal(path.Paths{np().Key("replicas")}, filtered.Unused)
}

func (s *Unittest) TestFilterIgnoredWithMerge_Undefined() {
	undefined := path.Paths{
		np().Key("a").Key("b").Key("c"),
		np().Key("a").Key("x").Key("c"),
		np().Key("a").Key("b").Key("d").Key("c"),
		np().Key("other"),
	}

	ignorePatterns := path.Paths{
		np().Key("a").Wildcard().Key("c"), // Interior wildcard matches one level
	}

	result := &Result{
		Referenced: path.Paths{np().Key("ref")},
		Unused:     path.Paths{},
		Undefined:  undefined,
	}

	filtered := filterIgnoredWithMerge(result, ignorePatterns)

	s.Require().Equal(result.Referenced, filtered.Referenced)
	expected := path.Paths{
		np().Key("a").Key("b").Key("d").Key("c"),
		np().Key("other"),
	}
	s.Require().Equal(expected, filtered.Undefined)
}

func (s *Unittest) TestFilterIgnoredWithMerge_MultiplePatterns() {
	unused := path.Paths{
		np().Key("image").Key("tag"),
		np().Key("replicas"),
		np().Key("config").Key("value"),
		np().Key("other").Key("field"),
	}

	ignorePatterns := path.Paths{
		np().Key("image").Key("tag"),
		np().Key("config").Wildcard(),
	}

	result := &Result{
		Referenced: path.Paths{np().Key("ref")},
		Unused:     unused,
		Undefined:  path.Paths{},
	}

	filtered := filterIgnoredWithMerge(result, ignorePatterns)

	expected := path.Paths{
		np().Key("other").Key("field"),
		np().Key("replicas"),
	}
	s.Require().Equal(expected, filtered.Unused)
}

func (s *Unittest) TestFilterIgnoredWithMerge_NoMatches() {
	unused := path.Paths{
		np().Key("image").Key("tag"),
		np().Key("replicas"),
	}

	ignorePatterns := path.Paths{
		np().Key("nonexistent"),
	}

	result := &Result{
		Referenced: path.Paths{np().Key("ref")},
		Unused:     unused,
		Undefined:  path.Paths{},
	}

	filtered := filterIgnoredWithMerge(result, ignorePatterns)

	s.Require().Equal(unused, filtered.Unused)
}

func (s *Unittest) TestFilterIgnoredWithMerge_EmptyIgnore() {
	unused := path.Paths{np().Key("image").Key("tag")}

	result := &Result{
		Referenced: path.Paths{np().Key("ref")},
		Unused:     unused,
		Undefined:  path.Paths{},
	}

	filtered := filterIgnoredWithMerge(result, path.Paths{})

	s.Require().Equal(unused, filtered.Unused)
}

func (s *Unittest) TestFilterIgnoredWithMerge_BothLists() {
	unused := path.Paths{
		np().Key("unused1"),
		np().Key("unused2"),
	}
	undefined := path.Paths{
		np().Key("undefined1"),
		np().Key("undefined2"),
	}

	ignorePatterns := path.Paths{
		np().Key("unused1"),
		np().Key("undefined1"),
	}

	result := &Result{
		Referenced: path.Paths{np().Key("ref")},
		Unused:     unused,
		Undefined:  undefined,
	}

	filtered := filterIgnoredWithMerge(result, ignorePatterns)

	s.Require().Equal(result.Referenced, filtered.Referenced)
	s.Require().Equal(path.Paths{np().Key("unused2")}, filtered.Unused)
	s.Require().Equal(path.Paths{np().Key("undefined2")}, filtered.Undefined)
}

func (s *Unittest) TestFilterIgnoredWithMerge_ReferencedNeverFiltered() {
	referenced := path.Paths{
		np().Key("ref1"),
		np().Key("ref2"),
	}

	ignorePatterns := path.Paths{
		np().Key("ref1"),
		np().Key("ref2"),
	}

	result := &Result{
		Referenced: referenced,
		Unused:     path.Paths{},
		Undefined:  path.Paths{},
	}

	filtered := filterIgnoredWithMerge(result, ignorePatterns)

	s.Require().Equal(referenced, filtered.Referenced)
}
