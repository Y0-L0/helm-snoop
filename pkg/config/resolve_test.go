package config

import (
	"os"
	"path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

func np() *vpath.Path { return &vpath.Path{} }

type testOptions struct {
	configPath  string
	noConfig    bool
	ignore      vpath.Paths
	valuesFiles []string
}

func (o *testOptions) ConfigPath() string    { return o.configPath }
func (o *testOptions) NoConfig() bool        { return o.noConfig }
func (o *testOptions) Ignore() vpath.Paths   { return o.ignore }
func (o *testOptions) ValuesFiles() []string { return o.valuesFiles }

func (s *Unittest) TestResolve_NoConfig() {
	charts, err := Resolve(
		[]string{"/charts/a", "/charts/b"},
		&testOptions{noConfig: true},
	)
	s.Require().NoError(err)
	s.Require().Len(charts, 2)
	s.Equal("/charts/a", charts[0].Path)
	s.Equal("/charts/b", charts[1].Path)
	s.Empty(charts[0].Ignore)
	s.Empty(charts[0].ValuesFiles)
	s.Nil(charts[0].ExtraValues)
}

func (s *Unittest) TestResolve_CLIFlagsOnly() {
	ignore := vpath.Paths{np().Key("image").Key("tag")}
	charts, err := Resolve(
		[]string{"/charts/a"},
		&testOptions{
			noConfig:    true,
			ignore:      ignore,
			valuesFiles: []string{"/extra/values.yaml"},
		},
	)
	s.Require().NoError(err)
	s.Require().Len(charts, 1)
	vpath.EqualPaths(s, ignore, charts[0].Ignore)
	s.Equal([]string{"/extra/values.yaml"}, charts[0].ValuesFiles)
}

func (s *Unittest) TestResolve_ConfigGlobalMergedWithCLI() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, ".helm-snoop.yaml")
	s.Require().NoError(os.WriteFile(configPath, []byte(`
version: 0
global:
  ignore:
    - .config.debug
  valuesFiles:
    - global-extra.yaml
  extraValues:
    globalKey: globalVal
`), 0o600))

	charts, err := Resolve(
		[]string{filepath.Join(dir, "my-chart")},
		&testOptions{
			configPath:  configPath,
			ignore:      vpath.Paths{np().Key("cli").Key("ignore")},
			valuesFiles: []string{"/cli/extra.yaml"},
		},
	)
	s.Require().NoError(err)
	s.Require().Len(charts, 1)

	// Global config ignore + CLI ignore.
	expectedIgnore := vpath.Paths{
		np().Key("config").Key("debug"),
		np().Key("cli").Key("ignore"),
	}
	vpath.EqualPaths(s, expectedIgnore, charts[0].Ignore)

	// Global config valuesFiles (resolved to configDir) + CLI valuesFiles.
	s.Equal(
		[]string{filepath.Join(dir, "global-extra.yaml"), "/cli/extra.yaml"},
		charts[0].ValuesFiles,
	)

	// Global extraValues passed through.
	s.Equal(map[string]any{"globalKey": "globalVal"}, charts[0].ExtraValues)
}

func (s *Unittest) TestResolve_PerChartOverlay() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, ".helm-snoop.yaml")
	s.Require().NoError(os.WriteFile(configPath, []byte(`
version: 0
global:
  ignore:
    - .global.path
  extraValues:
    shared: fromGlobal
    overridden: fromGlobal
charts:
  my-chart:
    ignore:
      - .chart.path
    valuesFiles:
      - chart-extra.yaml
    extraValues:
      overridden: fromChart
      chartOnly: val
`), 0o600))

	chartPath := filepath.Join(dir, "my-chart")
	charts, err := Resolve([]string{chartPath}, &testOptions{configPath: configPath})
	s.Require().NoError(err)
	s.Require().Len(charts, 1)

	// Global + per-chart ignore.
	expectedIgnore := vpath.Paths{
		np().Key("global").Key("path"),
		np().Key("chart").Key("path"),
	}
	vpath.EqualPaths(s, expectedIgnore, charts[0].Ignore)

	// Per-chart valuesFiles (resolved).
	s.Equal([]string{filepath.Join(dir, "chart-extra.yaml")}, charts[0].ValuesFiles)

	// ExtraValues: chart wins.
	s.Equal(map[string]any{
		"shared":     "fromGlobal",
		"overridden": "fromChart",
		"chartOnly":  "val",
	}, charts[0].ExtraValues)
}

func (s *Unittest) TestResolve_SkipPerChart() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, ".helm-snoop.yaml")
	s.Require().NoError(os.WriteFile(configPath, []byte(`
version: 0
charts:
  skipped-chart:
    skip: true
  normal-chart:
    ignore:
      - .foo
`), 0o600))

	charts, err := Resolve(
		[]string{
			filepath.Join(dir, "skipped-chart"),
			filepath.Join(dir, "normal-chart"),
			filepath.Join(dir, "unlisted-chart"),
		},
		&testOptions{configPath: configPath},
	)
	s.Require().NoError(err)
	s.Require().Len(charts, 3)
	s.True(charts[0].Skip, "skipped-chart should be skipped")
	s.False(charts[1].Skip, "normal-chart should not be skipped")
	s.False(charts[2].Skip, "unlisted-chart should not be skipped")
}

func (s *Unittest) TestResolve_UnmatchedChartIgnored() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, ".helm-snoop.yaml")
	s.Require().NoError(os.WriteFile(configPath, []byte(`
version: 0
charts:
  other-chart:
    ignore:
      - .something
`), 0o600))

	chartPath := filepath.Join(dir, "my-chart")
	charts, err := Resolve([]string{chartPath}, &testOptions{configPath: configPath})
	s.Require().NoError(err)
	s.Require().Len(charts, 1)
	s.Empty(charts[0].Ignore)
}

func (s *Unittest) TestResolve_ExplicitConfigNotFound() {
	_, err := Resolve(
		[]string{"/charts/a"},
		&testOptions{configPath: "/nonexistent/.helm-snoop.yaml"},
	)
	s.Require().Error(err)
	s.Contains(err.Error(), "reading config file")
}

func (s *Unittest) TestResolve_InvalidConfig() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, ".helm-snoop.yaml")
	s.Require().NoError(os.WriteFile(configPath, []byte(`version: 99`), 0o600))

	_, err := Resolve([]string{"/charts/a"}, &testOptions{configPath: configPath})
	s.Require().Error(err)
	s.Contains(err.Error(), "unsupported config version")
}

func (s *Unittest) TestResolve_ValuesFilesResolvedRelativeToConfig() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, ".helm-snoop.yaml")
	s.Require().NoError(os.WriteFile(configPath, []byte(`
version: 0
global:
  valuesFiles:
    - relative.yaml
    - /absolute/path.yaml
`), 0o600))

	charts, err := Resolve(
		[]string{filepath.Join(dir, "chart")},
		&testOptions{configPath: configPath},
	)
	s.Require().NoError(err)
	s.Equal([]string{
		filepath.Join(dir, "relative.yaml"),
		"/absolute/path.yaml",
	}, charts[0].ValuesFiles)
}

func (s *Unittest) TestResolve_NoConfigFileFound() {
	// No --config, no config file in the filesystem → empty options applied.
	charts, err := Resolve(
		[]string{"/charts/a"},
		&testOptions{},
	)
	s.Require().NoError(err)
	s.Require().Len(charts, 1)
	s.Equal("/charts/a", charts[0].Path)
	s.Empty(charts[0].Ignore)
}
