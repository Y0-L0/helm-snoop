package snooper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/y0-l0/helm-snoop/internal/testsuite"
)

type Unittest struct {
	testsuite.LoggingSuite
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }

// Integrationtest is the integration test suite.
// It tests with the real helm chart in the test-chart directory
type Integrationtest struct {
	testsuite.LoggingSuite
	chartPath string
}

func (s *Integrationtest) SetupSuite() {
	s.chartPath = "../../test-chart"
	_, err := os.Stat(s.chartPath)
	s.Require().NoError(err, "expected test chart at %s", s.chartPath)
}

func TestIntegration(t *testing.T) { suite.Run(t, new(Integrationtest)) }
