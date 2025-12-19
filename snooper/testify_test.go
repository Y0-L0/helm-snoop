package snooper

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// LoggingSuite sets a default slog logger per test for diagnostics.
type LoggingSuite struct {
	suite.Suite
	logBuf bytes.Buffer
}

func (s *LoggingSuite) SetupTest() {
	s.logBuf.Reset()
	handler := slog.NewTextHandler(&s.logBuf, &slog.HandlerOptions{AddSource: false, Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))
}

func (loggingSuite *LoggingSuite) TearDownTest() {
	if !loggingSuite.T().Failed() || !testing.Verbose() {
		return
	}
	loggingSuite.T().Log("=== Captured Production Logs ===\n")
	loggingSuite.T().Log(loggingSuite.logBuf.String())
}

type Unittest struct {
	LoggingSuite
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }

// Integrationtest is the integration test suite.
// It tests with the real helm chart in the test-chart directory
type Integrationtest struct {
	LoggingSuite
	chartPath string
}

func (s *Integrationtest) SetupSuite() {
	s.chartPath = "../test-chart"
	_, err := os.Stat(s.chartPath)
	s.Require().NoError(err, "expected test chart at %s", s.chartPath)
}

func TestIntegration(t *testing.T) { suite.Run(t, new(Integrationtest)) }
