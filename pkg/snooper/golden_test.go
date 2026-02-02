package snooper

import (
	"encoding/json"
	"path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/parser"
)

func (s *GoldenTest) TestSnoop_TestChart() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop(filepath.Join(s.chartsDir, "test-chart"), nil)
	s.Require().NoError(err)

	actual := results.toJSON()
	s.EqualGoldenJSON("test-chart.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_Chart_IntercomService() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop(filepath.Join(s.chartsDir, "intercom-service-2.23.0.tgz"), nil)
	s.Require().NoError(err)

	actual := results.toJSON()
	s.EqualGoldenJSON("intercom-service.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_Chart_Guardian() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop(filepath.Join(s.chartsDir, "guardian-0.24.4.tgz"), nil)
	s.Require().NoError(err)

	actual := results.toJSON()
	s.EqualGoldenJSON("guardian.results.golden.json", actual)
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
	s.Equal(expected.Referenced, actual.Referenced, "Referenced paths mismatch")
	s.Equal(expected.Undefined, actual.Undefined, "Undefined paths mismatch")
	s.Equal(expected.Unused, actual.Unused, "Unused paths mismatch")
}

func disableStrictParsing() func() {
	oldStrict := parser.Strict
	parser.Strict = false

	return func() { parser.Strict = oldStrict }
}
