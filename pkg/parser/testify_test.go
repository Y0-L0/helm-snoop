package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/y0-l0/helm-snoop/internal/testsuite"
	"github.com/y0-l0/helm-snoop/pkg/path"
)

type Unittest struct {
	testsuite.LoggingSuite
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }

func (s *Unittest) parse(tmpl string) path.Paths {
	paths, err := parseFile("", "test.yaml", []byte(tmpl), nil)
	s.Require().NoError(err)
	return paths
}
