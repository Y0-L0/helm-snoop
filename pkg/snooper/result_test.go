package snooper

import (
	"github.com/y0-l0/helm-snoop/pkg/vpath"
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
				Unused:    vpath.Paths{vpath.NewPath("unused")},
				Undefined: vpath.Paths{},
			},
			true,
		},
		{
			"has Undefined findings",
			&Result{
				Unused:    vpath.Paths{},
				Undefined: vpath.Paths{vpath.NewPath("undefined")},
			},
			true,
		},
		{
			"has both types of findings",
			&Result{
				Unused:    vpath.Paths{vpath.NewPath("unused")},
				Undefined: vpath.Paths{vpath.NewPath("undefined")},
			},
			true,
		},
		{
			"no findings",
			&Result{
				Unused:    vpath.Paths{},
				Undefined: vpath.Paths{},
			},
			false,
		},
		{
			"only Referenced (no findings)",
			&Result{
				Referenced: vpath.Paths{vpath.NewPath("used")},
				Unused:     vpath.Paths{},
				Undefined:  vpath.Paths{},
			},
			false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := tc.result.hasFindings()
			s.Equal(tc.want, got)
		})
	}
}
