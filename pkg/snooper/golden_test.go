package snooper

import (
	"encoding/json"
	"path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/parser"
	loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

func (s *GoldenTest) TestSnoop_TestChart() {
	chart, err := loader.Load(filepath.Join(s.chartsDir, "test-chart"))
	s.Require().NoError(err)

	results, err := Snoop(chart)
	s.Require().NoError(err)

	actual := results.ToJSON()
	s.EqualGoldenJSON("test-chart.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_IntercomService() {
	restore := disableStrictParsing()
	defer restore()

	chart, err := loader.Load(filepath.Join(s.chartsDir, "intercom-service-2.23.0.tgz"))
	s.Require().NoError(err)

	results, err := Snoop(chart)
	s.Require().NoError(err)

	actual := results.ToJSON()
	s.EqualGoldenJSON("intercom-service.results.golden.json", actual)
}

func (s *GoldenTest) EqualGoldenJSON(name string, actual ResultsJSON) {
	path := s.goldenPath(name)
	if s.update {
		data, err := json.MarshalIndent(actual, "", "  ")
		s.Require().NoError(err)
		s.WriteFile(path, data)
	}
	var expected ResultsJSON
	s.Require().NoError(json.Unmarshal(s.ReadFile(path), &expected))

	// Compare each field separately for clearer diffs
	s.Require().Equal(expected.Referenced, actual.Referenced, "Referenced paths mismatch")
	s.Require().Equal(expected.UsedNotDefined, actual.UsedNotDefined, "UsedNotDefined paths mismatch")
	s.Require().Equal(expected.DefinedNotUsed, actual.DefinedNotUsed, "DefinedNotUsed paths mismatch")
}

func disableStrictParsing() func() {
	oldStrict := parser.Strict
	parser.Strict = false

	return func() { parser.Strict = oldStrict }
}
