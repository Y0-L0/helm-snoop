package parser

import (
	"log/slog"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/y0-l0/helm-snoop/internal/testsuite"
	"github.com/y0-l0/helm-snoop/snooper/path"
)

type Unittest struct {
	testsuite.LoggingSuite
}

func (s *Unittest) EqualPaths(expected path.Paths, actual path.Paths) {
	slog.Debug("Asserting equal Paths", "expected", expected, "actual", actual)
	s.Equal(len(expected), len(actual))
	sort.Sort(expected)
	sort.Sort(actual)

	for i, exp := range expected {
		act := &path.Path{}
		if i < len(actual) {
			act = actual[i]
		}
		s.Equal(exp, act)
	}
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }
