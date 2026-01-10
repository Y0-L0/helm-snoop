package snooper

import (
	"bytes"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

func (s *Unittest) TestHasFindings() {
	tests := []struct {
		name   string
		result *Result
		want   bool
	}{
		{
			"has DefinedNotUsed findings",
			&Result{
				DefinedNotUsed: path.Paths{path.NewPath("unused")},
				UsedNotDefined: path.Paths{},
			},
			true,
		},
		{
			"has UsedNotDefined findings",
			&Result{
				DefinedNotUsed: path.Paths{},
				UsedNotDefined: path.Paths{path.NewPath("undefined")},
			},
			true,
		},
		{
			"has both types of findings",
			&Result{
				DefinedNotUsed: path.Paths{path.NewPath("unused")},
				UsedNotDefined: path.Paths{path.NewPath("undefined")},
			},
			true,
		},
		{
			"no findings",
			&Result{
				DefinedNotUsed: path.Paths{},
				UsedNotDefined: path.Paths{},
			},
			false,
		},
		{
			"only Referenced (no findings)",
			&Result{
				Referenced:     path.Paths{path.NewPath("used")},
				DefinedNotUsed: path.Paths{},
				UsedNotDefined: path.Paths{},
			},
			false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := tc.result.HasFindings()
			s.Equal(tc.want, got)
		})
	}
}

func (s *Unittest) TestToText() {
	tests := []struct {
		name   string
		result *Result
		want   []string
	}{
		{
			"all sections populated",
			&Result{
				Referenced:     path.Paths{path.NewPath("ref1"), path.NewPath("ref2")},
				DefinedNotUsed: path.Paths{path.NewPath("unused1")},
				UsedNotDefined: path.Paths{path.NewPath("undefined1")},
			},
			[]string{
				"Referenced:",
				"  /ref1",
				"  /ref2",
				"Defined-not-used:",
				"  /unused1",
				"Used-not-defined:",
				"  /undefined1",
			},
		},
		{
			"empty result",
			&Result{
				Referenced:     path.Paths{},
				DefinedNotUsed: path.Paths{},
				UsedNotDefined: path.Paths{},
			},
			[]string{
				"Referenced:",
				"Defined-not-used:",
				"Used-not-defined:",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var buf bytes.Buffer
			err := tc.result.ToText(&buf)
			s.Require().NoError(err)

			output := buf.String()
			for _, expected := range tc.want {
				s.Contains(output, expected)
			}
		})
	}
}
