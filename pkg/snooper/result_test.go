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
			"has Unused findings",
			&Result{
				Unused:    path.Paths{path.NewPath("unused")},
				Undefined: path.Paths{},
			},
			true,
		},
		{
			"has Undefined findings",
			&Result{
				Unused:    path.Paths{},
				Undefined: path.Paths{path.NewPath("undefined")},
			},
			true,
		},
		{
			"has both types of findings",
			&Result{
				Unused:    path.Paths{path.NewPath("unused")},
				Undefined: path.Paths{path.NewPath("undefined")},
			},
			true,
		},
		{
			"no findings",
			&Result{
				Unused:    path.Paths{},
				Undefined: path.Paths{},
			},
			false,
		},
		{
			"only Referenced (no findings)",
			&Result{
				Referenced: path.Paths{path.NewPath("used")},
				Unused:     path.Paths{},
				Undefined:  path.Paths{},
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
				Referenced: path.Paths{path.NewPath("ref1"), path.NewPath("ref2")},
				Unused:     path.Paths{path.NewPath("unused1")},
				Undefined:  path.Paths{path.NewPath("undefined1")},
			},
			[]string{
				"Referenced:",
				"  .ref1",
				"  .ref2",
				"Unused:",
				"  .unused1",
				"Undefined:",
				"  .undefined1",
			},
		},
		{
			"empty result",
			&Result{
				Referenced: path.Paths{},
				Unused:     path.Paths{},
				Undefined:  path.Paths{},
			},
			[]string{
				"Referenced:",
				"Unused:",
				"Undefined:",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var buf bytes.Buffer
			err := tc.result.ToText(&buf, true)
			s.Require().NoError(err)

			output := buf.String()
			for _, expected := range tc.want {
				s.Contains(output, expected)
			}
		})
	}
}
