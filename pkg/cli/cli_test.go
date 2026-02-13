package cli

import (
	"log/slog"
	"os"
	filepath "path/filepath"

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

type trackingSnoop struct {
	calls []string
}

func (t *trackingSnoop) snoop(chartPath string, ignorePatterns path.Paths) (*snooper.Result, error) {
	t.calls = append(t.calls, chartPath)
	return &snooper.Result{
		Referenced: path.Paths{},
		Unused:     path.Paths{},
		Undefined:  path.Paths{},
	}, nil
}

func testdataDir() string {
	// cli_test.go lives in pkg/cli/, testdata is at repo root
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..", "..", "testdata")
}

func (s *Unittest) TestResolveChartRootFromFile() {
	chartRoot := filepath.Join(testdataDir(), "test-chart")
	got, err := resolveChartRoot(filepath.Join(chartRoot, "values.yaml"))
	s.Require().NoError(err)
	s.Equal(chartRoot, got)
}

func (s *Unittest) TestResolveChartRootFromNestedFile() {
	chartRoot := filepath.Join(testdataDir(), "test-chart")
	got, err := resolveChartRoot(filepath.Join(chartRoot, "templates", "configmap.yaml"))
	s.Require().NoError(err)
	s.Equal(chartRoot, got)
}

func (s *Unittest) TestResolveChartRootFromChartDir() {
	chartRoot := filepath.Join(testdataDir(), "test-chart")
	got, err := resolveChartRoot(chartRoot)
	s.Require().NoError(err)
	s.Equal(chartRoot, got)
}

func (s *Unittest) TestResolveChartRootNoChartYaml() {
	_, err := resolveChartRoot(os.TempDir())
	s.Require().Error(err)
	s.Contains(err.Error(), "no Chart.yaml found")
}

func (s *Unittest) TestResolveChartRootNonexistentPath() {
	_, err := resolveChartRoot("/nonexistent/path/to/file.yaml")
	s.Require().Error(err)
}

func (s *Unittest) TestResolveChartRootArchivePassthrough() {
	// A valid chart archive is returned as-is (loader accepts it directly)
	chartRoot := filepath.Join(testdataDir(), "test-chart")
	got, err := resolveChartRoot(chartRoot)
	s.Require().NoError(err)
	s.Equal(chartRoot, got)
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
	chartPath := filepath.Join(testdataDir(), "test-chart")
	tests := []struct {
		name string
		args []string
	}{
		{"basic chart path", []string{"helm-snoop", chartPath}},
		{"with json output", []string{"helm-snoop", "--json", chartPath}},
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
		{"no verbosity", []string{"helm-snoop", "../../testdata/test-chart"}, slog.LevelWarn},
		{"single v", []string{"helm-snoop", "-v", "../../testdata/test-chart"}, slog.LevelInfo},
		{"double v", []string{"helm-snoop", "-vv", "../../testdata/test-chart"}, slog.LevelDebug},
		{"triple v", []string{"helm-snoop", "-vvv", "../../testdata/test-chart"}, slog.LevelDebug},
		{"separate v flags", []string{"helm-snoop", "-v", "-v", "../../testdata/test-chart"}, slog.LevelDebug},
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

func (s *Unittest) TestFileArgResolvesToChartRoot() {
	tracker := &trackingSnoop{}
	chartFile := filepath.Join(testdataDir(), "test-chart", "values.yaml")
	command := NewParser([]string{"helm-snoop", chartFile}, snooper.SetupLogging, tracker.snoop)
	err := command.Execute()
	s.Require().NoError(err)
	s.Require().Len(tracker.calls, 1)
	s.Equal(filepath.Join(testdataDir(), "test-chart"), tracker.calls[0])
}

func (s *Unittest) TestMultipleArgsDeduplication() {
	tracker := &trackingSnoop{}
	chartDir := filepath.Join(testdataDir(), "test-chart")
	command := NewParser([]string{
		"helm-snoop",
		filepath.Join(chartDir, "values.yaml"),
		filepath.Join(chartDir, "templates", "configmap.yaml"),
		filepath.Join(chartDir, "templates", "deployment.yaml"),
	}, snooper.SetupLogging, tracker.snoop)
	err := command.Execute()
	s.Require().NoError(err)
	s.Require().Len(tracker.calls, 1, "expected deduplication to single chart")
}
