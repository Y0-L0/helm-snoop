package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/y0-l0/helm-snoop/internal/testsuite"
)

var update = flag.Bool("update", false, "update golden files")

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

func (s *GoldenTest) goldenFile(name string) string { return filepath.Join("testdata", name) }
