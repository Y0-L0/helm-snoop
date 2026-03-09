package config

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/y0-l0/helm-snoop/internal/testsuite"
)

type Unittest struct {
	testsuite.LoggingSuite
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }

func (s *Unittest) TestParse_Valid() {
	data := []byte(`
version: 0
global:
  ignore:
    - .image.tag
  valuesFiles:
    - extra.yaml
  extraValues:
    key: value
charts:
  my-chart:
    ignore:
      - .debug
    valuesFiles:
      - override.yaml
    extraValues:
      nested:
        key: val
`)
	cfg, err := parse(data)
	s.Require().NoError(err)

	s.Equal([]string{".image.tag"}, cfg.Global.Ignore)
	s.Equal([]string{"extra.yaml"}, cfg.Global.ValuesFiles)
	s.Equal(map[string]any{"key": "value"}, cfg.Global.ExtraValues)

	chart, ok := cfg.Charts["my-chart"]
	s.Require().True(ok)
	s.Equal([]string{".debug"}, chart.Ignore)
	s.Equal([]string{"override.yaml"}, chart.ValuesFiles)
	s.Equal(map[string]any{"nested": map[string]any{"key": "val"}}, chart.ExtraValues)
}

func (s *Unittest) TestParse_MissingVersion() {
	data := []byte(`
global:
  ignore:
    - .foo
`)
	_, err := parse(data)
	s.Require().Error(err)
	s.Contains(err.Error(), "missing required 'version'")
}

func (s *Unittest) TestParse_UnsupportedVersion() {
	data := []byte(`version: 99`)
	_, err := parse(data)
	s.Require().Error(err)
	s.Contains(err.Error(), "unsupported config version 99")
}

func (s *Unittest) TestParse_InvalidYAML() {
	data := []byte(`{{{not yaml`)
	_, err := parse(data)
	s.Require().Error(err)
	s.Contains(err.Error(), "parsing config file")
}

func (s *Unittest) TestParse_MinimalValid() {
	data := []byte(`version: 0`)
	cfg, err := parse(data)
	s.Require().NoError(err)
	s.Empty(cfg.Global.Ignore)
	s.Empty(cfg.Global.ValuesFiles)
	s.Empty(cfg.Global.ExtraValues)
	s.Empty(cfg.Charts)
}

func (s *Unittest) TestParse_Skip() {
	data := []byte(`
version: 0
charts:
  my-chart:
    skip: true
  other-chart:
    ignore:
      - .foo
`)
	cfg, err := parse(data)
	s.Require().NoError(err)

	s.True(cfg.Charts["my-chart"].Skip)
	s.False(cfg.Charts["other-chart"].Skip)
}

func (s *Unittest) TestParse_GlobalOnly() {
	data := []byte(`
version: 0
global:
  ignore:
    - .a
    - .b
`)
	cfg, err := parse(data)
	s.Require().NoError(err)
	s.Equal([]string{".a", ".b"}, cfg.Global.Ignore)
	s.Empty(cfg.Charts)
}
