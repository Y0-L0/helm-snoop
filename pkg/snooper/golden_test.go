package snooper

import (
	"encoding/json"
	"path/filepath"

	"github.com/y0-l0/helm-snoop/internal/assert"
)

func (s *GoldenTest) TestSnoop_json_TestChart() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop(filepath.Join(s.chartsDir, "test-chart"), nil)
	s.Require().NoError(err)

	actual := results.toJSON()
	s.EqualGoldenJSON("test-chart.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_json_IntercomService() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop(filepath.Join(s.chartsDir, "intercom-service-2.23.0.tgz"), nil)
	s.Require().NoError(err)

	actual := results.toJSON()
	s.EqualGoldenJSON("intercom-service.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_json_Guardian() {
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

func (s *GoldenTest) TestSnoop_txt_TestChart() {
	restore := disableStrictParsing()
	defer restore()

	result, err := Snoop(filepath.Join(s.chartsDir, "test-chart"), nil)
	s.Require().NoError(err)
	s.compactGoldenTest("test-chart.results", Results{result})
}

func (s *GoldenTest) TestSnoop_txt_IntercomService() {
	restore := disableStrictParsing()
	defer restore()

	result, err := Snoop(filepath.Join(s.chartsDir, "intercom-service-2.23.0.tgz"), nil)
	s.Require().NoError(err)
	s.compactGoldenTest("intercom-service.results", Results{result})
}

func (s *GoldenTest) TestSnoop_txt_Guardian() {
	restore := disableStrictParsing()
	defer restore()

	result, err := Snoop(filepath.Join(s.chartsDir, "guardian-0.24.4.tgz"), nil)
	s.Require().NoError(err)
	s.compactGoldenTest("guardian.results", Results{result})
}

func (s *GoldenTest) TestSnoop_UnusedHaveValuesContext() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop(filepath.Join(s.chartsDir, "test-chart"), nil)
	s.Require().NoError(err)

	for _, p := range results.Unused {
		s.Require().NotEmpty(p.Contexts, "unused path %s should have a context", p.ID())
		s.Equal("values.yaml", p.Contexts[0].FileName,
			"unused path %s context should point to values.yaml", p.ID())
		s.Greater(p.Contexts[0].Line, 0,
			"unused path %s should have a positive line number", p.ID())
	}
}

func disableStrictParsing() func() {
	oldStrict := assert.Strict
	assert.Strict = false

	return func() { assert.Strict = oldStrict }
}
