package cli

import (
	"bytes"
	"log/slog"
	"os"
	filepath "path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/snooper"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

func mockSnoop(charts snooper.Charts) error {
	for _, c := range charts {
		c.Name = filepath.Base(c.Path)
		c.Result = &snooper.Result{}
	}
	return nil
}

type trackingSnoop struct {
	calls []string
}

func (t *trackingSnoop) snoop(charts snooper.Charts) error {
	for _, c := range charts {
		t.calls = append(t.calls, c.Path)
		c.Name = filepath.Base(c.Path)
		c.Result = &snooper.Result{}
	}
	return nil
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
		{"no verbosity", []string{"helm-snoop", "../../testdata/test-chart"}, slog.LevelError},
		{"single v", []string{"helm-snoop", "-v", "../../testdata/test-chart"}, slog.LevelWarn},
		{"double v", []string{"helm-snoop", "-vv", "../../testdata/test-chart"}, slog.LevelInfo},
		{"triple v", []string{"helm-snoop", "-vvv", "../../testdata/test-chart"}, slog.LevelDebug},
		{"separate v flags", []string{"helm-snoop", "-v", "-v", "../../testdata/test-chart"}, slog.LevelInfo},
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

func (s *Unittest) TestMultipleChartsSingleSummary() {
	chartA := filepath.Join(testdataDir(), "test-chart")
	chartB := s.T().TempDir()
	err := os.WriteFile(
		filepath.Join(chartB, "Chart.yaml"),
		[]byte("name: temp-chart"),
		0o600,
	)
	s.Require().NoError(err)

	findingsSnoop := func(charts snooper.Charts) error {
		for _, c := range charts {
			unused := vpath.NewPath("someUnused")
			unused.Contexts = vpath.Contexts{{FileName: "values.yaml", Line: 1, Column: 1}}
			c.Name = filepath.Base(c.Path)
			c.Result = &snooper.Result{
				Unused: vpath.Paths{unused},
			}
		}
		return nil
	}

	var stdout, stderr bytes.Buffer
	code := Main(
		[]string{"helm-snoop", chartA, chartB},
		&stdout, &stderr,
		snooper.SetupLogging, findingsSnoop,
	)

	s.Equal(1, code, "expected exit code 1 for findings")
	out := stdout.String()
	s.Contains(out, "2 charts")
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

func (s *Unittest) TestConfigFlag() {
	var captured snooper.Charts

	mockSnoop := func(charts snooper.Charts) error {
		captured = charts
		for _, c := range charts {
			c.Name = filepath.Base(c.Path)
			c.Result = &snooper.Result{}
		}
		return nil
	}

	dir := s.T().TempDir()
	configPath := filepath.Join(dir, ".helm-snoop.yaml")
	s.Require().NoError(os.WriteFile(configPath, []byte(`
version: 0
global:
  ignore:
    - .from.config
`), 0o600))

	chartPath := filepath.Join(testdataDir(), "test-chart")
	command := NewParser(
		[]string{"helm-snoop", "--config", configPath, chartPath},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)
	s.Require().Len(captured, 1)

	var ids []string
	for _, p := range captured[0].Ignore {
		ids = append(ids, p.ID())
	}
	s.Contains(ids, ".from.config")
}

func (s *Unittest) TestNoConfigFlag() {
	var captured snooper.Charts

	mockSnoop := func(charts snooper.Charts) error {
		captured = charts
		for _, c := range charts {
			c.Name = filepath.Base(c.Path)
			c.Result = &snooper.Result{}
		}
		return nil
	}

	chartPath := filepath.Join(testdataDir(), "test-chart")
	command := NewParser(
		[]string{"helm-snoop", "--no-config", chartPath},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)
	s.Require().Len(captured, 1)
	s.Empty(captured[0].Ignore)
	s.Empty(captured[0].ValuesFiles)
	s.Nil(captured[0].ExtraValues)
}

func (s *Unittest) TestConfigAndCLIFlagsMerge() {
	var captured snooper.Charts

	mockSnoop := func(charts snooper.Charts) error {
		captured = charts
		for _, c := range charts {
			c.Name = filepath.Base(c.Path)
			c.Result = &snooper.Result{}
		}
		return nil
	}

	dir := s.T().TempDir()
	configPath := filepath.Join(dir, ".helm-snoop.yaml")
	s.Require().NoError(os.WriteFile(configPath, []byte(`
version: 0
global:
  ignore:
    - .from.config
`), 0o600))

	chartPath := filepath.Join(testdataDir(), "test-chart")
	command := NewParser(
		[]string{
			"helm-snoop",
			"--config", configPath,
			"-i", ".from.cli",
			chartPath,
		},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)
	s.Require().Len(captured, 1)

	var ids []string
	for _, p := range captured[0].Ignore {
		ids = append(ids, p.ID())
	}
	s.Contains(ids, ".from.config")
	s.Contains(ids, ".from.cli")
}
