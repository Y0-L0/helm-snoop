package testsuite

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
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

// WriteFile creates parent directories as needed and writes data to path.
// Test fails on error.
func (s *LoggingSuite) WriteFile(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		s.Require().NoError(err, "mkdir parent for %s", path)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		s.Require().NoError(err, "write file %s", path)
	}
}

// ReadFile reads and returns file contents. Test fails on error.
func (s *LoggingSuite) ReadFile(path string) []byte {
	b, err := os.ReadFile(path)
	s.Require().NoError(err, "read file %s", path)
	return b
}
