package cli

import (
	"log/slog"

	"github.com/y0-l0/helm-snoop/pkg/path"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
	chart "helm.sh/helm/v4/pkg/chart/v2"
)

func mockLoader(path string) (*chart.Chart, error) {
	return &chart.Chart{}, nil
}

func mockSnoop(c *chart.Chart, ignore []string) (*snooper.Result, error) {
	return &snooper.Result{
		Referenced:     path.Paths{},
		DefinedNotUsed: path.Paths{},
		UsedNotDefined: path.Paths{},
	}, nil
}

func (s *Unittest) TestHelp() {
	command := NewParser([]string{"helm-snoop", "--help"}, snooper.SetupLogging)
	err := command.Execute()
	s.Require().NoError(err)
}

func (s *Unittest) TestVersionSubcommand() {
	command := NewParser([]string{"helm-snoop", "version"}, snooper.SetupLogging)
	err := command.Execute()
	s.Require().NoError(err)
}

func (s *Unittest) TestValidArguments() {
	tests := []struct {
		name string
		args []string
	}{
		{"basic chart path", []string{"helm-snoop", "/path/to/chart"}},
		{"with single ignore flag", []string{"helm-snoop", "-i", "image.tag", "/path/to/chart"}},
		{"with multiple ignore flags", []string{"helm-snoop", "-i", "image.tag", "-i", "replicas", "/path/to/chart"}},
		{"with json output", []string{"helm-snoop", "--json", "/path/to/chart"}},
		{"combined flags", []string{"helm-snoop", "-i", "tag", "--json", "/path/to/chart"}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			command := NewParserForTest(tc.args, snooper.SetupLogging, mockLoader, mockSnoop)
			err := command.Execute()
			s.NoError(err)
		})
	}
}

func (s *Unittest) TestInvalidArguments() {
	command := NewParser([]string{"helm-snoop"}, snooper.SetupLogging)
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

			command := NewParserForTest(tc.args, mockSetupLogging, mockLoader, mockSnoop)
			_ = command.Execute()

			s.Equal(tc.expected, capturedLevel)
		})
	}
}
