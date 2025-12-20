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
