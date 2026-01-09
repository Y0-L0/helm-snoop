package snooper

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/y0-l0/helm-snoop/internal/testsuite"
)

type Unittest struct {
	testsuite.LoggingSuite
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }

// GoldenTest is the golden test suite (CLI and Results goldens live elsewhere).
type GoldenTest struct {
	testsuite.LoggingSuite
	chartsDir string
	update    bool
}

func (s *GoldenTest) SetupSuite() {
	s.chartsDir = filepath.Join("..", "..", "testdata")
	s.update = *update
}

func TestGolden(t *testing.T) { suite.Run(t, new(GoldenTest)) }

var update = flag.Bool("update", false, "update golden files")

func (s *GoldenTest) goldenPath(name string) string { return filepath.Join("testdata", name) }
