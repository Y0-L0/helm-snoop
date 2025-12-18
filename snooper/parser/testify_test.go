package parser

import (
	"bytes"
	"log/slog"
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

type Unittest struct {
	LoggingSuite
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }
