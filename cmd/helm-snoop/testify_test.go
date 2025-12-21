package main

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/y0-l0/helm-snoop/internal/testsuite"
)

var update = flag.Bool("update", false, "update golden files")

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

func (s *GoldenTest) goldenFile(name string) string { return filepath.Join("testdata", name) }
