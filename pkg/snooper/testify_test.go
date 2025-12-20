package snooper

import (
	"flag"
	"os"
	"path/filepath"
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
	s.chartPath = "../../testdata/test-chart"
	_, err := os.Stat(s.chartPath)
	s.Require().NoError(err, "expected test chart at %s", s.chartPath)
}

func TestIntegration(t *testing.T) { suite.Run(t, new(Integrationtest)) }

// GoldenTest is the golden test suite (CLI and Results goldens live elsewhere).
type GoldenTest struct {
	testsuite.LoggingSuite
	chartPath string
	update    bool
}

func (s *GoldenTest) SetupSuite() {
	s.chartPath = filepath.Join("..", "..", "testdata", "test-chart")
	_, err := os.Stat(s.chartPath)
	s.Require().NoError(err, "expected test chart at %s", s.chartPath)
	s.update = *update
}

func TestGolden(t *testing.T) { suite.Run(t, new(GoldenTest)) }

var update = flag.Bool("update", false, "update golden files")

// helper methods for golden handling
func (s *GoldenTest) goldenPath(name string) string { return filepath.Join("testdata", name) }
func (s *GoldenTest) writeFile(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		s.Require().NoError(err, "mkdir testdata")
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		s.Require().NoError(err, "write golden")
	}
}
func (s *GoldenTest) readFile(path string) []byte {
	b, err := os.ReadFile(path)
	s.Require().NoError(err, "read golden")
	return b
}
