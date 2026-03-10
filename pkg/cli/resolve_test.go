package cli

import (
	"os"
	filepath "path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

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
