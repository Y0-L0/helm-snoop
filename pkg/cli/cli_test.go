package cli

import (
	"log/slog"

	"github.com/y0-l0/helm-snoop/pkg/path"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

func mockSnoop(chartPath string, ignorePatterns path.Paths) (*snooper.Result, error) {
	return &snooper.Result{
		Referenced: path.Paths{},
		Unused:     path.Paths{},
		Undefined:  path.Paths{},
	}, nil
}

func (s *Unittest) TestHelp() {
	command := NewParser([]string{"helm-snoop", "--help"}, snooper.SetupLogging, mockSnoop)
	err := command.Execute()
	s.Require().NoError(err)
}

func (s *Unittest) TestVersionSubcommand() {
	command := NewParser([]string{"helm-snoop", "version"}, snooper.SetupLogging, mockSnoop)
	err := command.Execute()
	s.Require().NoError(err)
}

func (s *Unittest) TestValidArguments() {
	tests := []struct {
		name string
		args []string
	}{
		{"basic chart path", []string{"helm-snoop", "/path/to/chart"}},
		{"with json output", []string{"helm-snoop", "--json", "/path/to/chart"}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			command := NewParser(tc.args, snooper.SetupLogging, mockSnoop)
			err := command.Execute()
			s.NoError(err)
		})
	}
}

func (s *Unittest) TestInvalidArguments() {
	command := NewParser([]string{"helm-snoop"}, snooper.SetupLogging, mockSnoop)
	err := command.Execute()
	s.Require().Error(err)
}

func (s *Unittest) TestVerbosityLevels() {
	tests := []struct {
		name     string
		args     []string
		expected slog.Level
	}{
		{"no verbosity", []string{"helm-snoop", "/path"}, slog.LevelWarn},
		{"single v", []string{"helm-snoop", "-v", "/path"}, slog.LevelInfo},
		{"double v", []string{"helm-snoop", "-vv", "/path"}, slog.LevelDebug},
		{"triple v", []string{"helm-snoop", "-vvv", "/path"}, slog.LevelDebug},
		{"separate v flags", []string{"helm-snoop", "-v", "-v", "/path"}, slog.LevelDebug},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var capturedLevel slog.Level
			mockSetupLogging := func(level slog.Level) {
				capturedLevel = level
			}

			command := NewParser(tc.args, mockSetupLogging, mockSnoop)
			_ = command.Execute()

			s.Equal(tc.expected, capturedLevel)
		})
	}
}
