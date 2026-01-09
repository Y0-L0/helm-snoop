package snooper

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
)

func (s *Unittest) TestFilterIgnored_NoopCases() {
	input := &Result{
		Referenced:     path.Paths{path.NewPath("config", "enabled")},
		DefinedNotUsed: path.Paths{path.NewPath("image", "tag")},
		UsedNotDefined: path.Paths{path.NewPath("missing", "value")},
	}

	tests := []struct {
		name   string
		ignore []string
	}{
		{"no ignore keys", []string{}},
		{"key not in result", []string{"/nonexistent"}},
		{"referenced list unchanged", []string{"/config/enabled"}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual := filterIgnored(input, tc.ignore)

			s.Equal(path.Paths{path.NewPath("config", "enabled")}, actual.Referenced)
			s.Equal(path.Paths{path.NewPath("image", "tag")}, actual.DefinedNotUsed)
			s.Equal(path.Paths{path.NewPath("missing", "value")}, actual.UsedNotDefined)
		})
	}
}

func (s *Unittest) TestFilterIgnored_SingleKeyInDefinedNotUsed() {
	input := &Result{
		Referenced:     path.Paths{path.NewPath("config", "enabled")},
		DefinedNotUsed: path.Paths{path.NewPath("image", "tag"), path.NewPath("replicas")},
		UsedNotDefined: path.Paths{path.NewPath("missing", "value")},
	}

	actual := filterIgnored(input, []string{"/image/tag"})

	s.Equal(path.Paths{path.NewPath("config", "enabled")}, actual.Referenced)
	s.Equal(path.Paths{path.NewPath("replicas")}, actual.DefinedNotUsed)
	s.Equal(path.Paths{path.NewPath("missing", "value")}, actual.UsedNotDefined)
}

func (s *Unittest) TestFilterIgnored_SingleKeyInUsedNotDefined() {
	input := &Result{
		Referenced:     path.Paths{path.NewPath("config", "enabled")},
		DefinedNotUsed: path.Paths{path.NewPath("image", "tag")},
		UsedNotDefined: path.Paths{path.NewPath("missing", "value"), path.NewPath("other")},
	}

	actual := filterIgnored(input, []string{"/missing/value"})

	s.Equal(path.Paths{path.NewPath("config", "enabled")}, actual.Referenced)
	s.Equal(path.Paths{path.NewPath("image", "tag")}, actual.DefinedNotUsed)
	s.Equal(path.Paths{path.NewPath("other")}, actual.UsedNotDefined)
}

func (s *Unittest) TestFilterIgnored_MultipleKeysFromBothLists() {
	input := &Result{
		Referenced:     path.Paths{path.NewPath("config", "enabled")},
		DefinedNotUsed: path.Paths{path.NewPath("image", "tag"), path.NewPath("replicas")},
		UsedNotDefined: path.Paths{path.NewPath("missing", "value"), path.NewPath("other")},
	}

	actual := filterIgnored(input, []string{"/image/tag", "/missing/value"})

	s.Equal(path.Paths{path.NewPath("config", "enabled")}, actual.Referenced)
	s.Equal(path.Paths{path.NewPath("replicas")}, actual.DefinedNotUsed)
	s.Equal(path.Paths{path.NewPath("other")}, actual.UsedNotDefined)
}

func (s *Unittest) TestFilterIgnored_AllFindings() {
	input := &Result{
		Referenced:     path.Paths{path.NewPath("config", "enabled")},
		DefinedNotUsed: path.Paths{path.NewPath("image", "tag")},
		UsedNotDefined: path.Paths{path.NewPath("missing", "value")},
	}

	actual := filterIgnored(input, []string{"/image/tag", "/missing/value"})

	s.Equal(path.Paths{path.NewPath("config", "enabled")}, actual.Referenced)
	s.Equal(path.Paths{}, actual.DefinedNotUsed)
	s.Equal(path.Paths{}, actual.UsedNotDefined)
}
