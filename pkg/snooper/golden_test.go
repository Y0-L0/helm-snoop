package snooper

import (
	"encoding/json"
	"path/filepath"

	"github.com/y0-l0/helm-snoop/internal/assert"
)

func (s *GoldenTest) TestSnoop_json_TestChart() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "test-chart")}})
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	actual := results[0].toJSON()
	s.EqualGoldenJSON("test-chart.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_json_IntercomService() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "intercom-service-2.23.0.tgz")}})
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	actual := results[0].toJSON()
	s.EqualGoldenJSON("intercom-service.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_json_Guardian() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "guardian-0.24.4.tgz")}})
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	actual := results[0].toJSON()
	s.EqualGoldenJSON("guardian.results.golden.json", actual)
}

func (s *GoldenTest) EqualGoldenJSON(name string, actual ResultJSON) {
	path := s.goldenPath(name)
	if s.update {
		data, err := json.MarshalIndent(actual, "", "  ")
		s.Require().NoError(err)
		s.WriteFile(path, data)
	}
	var expected ResultJSON
	s.Require().NoError(json.Unmarshal(s.ReadFile(path), &expected))

	// Compare each field separately for clearer diffs
	s.Equal(expected.Referenced, actual.Referenced, "Referenced paths mismatch")
	s.Equal(expected.Undefined, actual.Undefined, "Undefined paths mismatch")
	s.Equal(expected.Unused, actual.Unused, "Unused paths mismatch")
}

func (s *GoldenTest) TestSnoop_txt_TestChart() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "test-chart")}})
	s.Require().NoError(err)
	s.compactGoldenTest("test-chart.results", results)
}

func (s *GoldenTest) TestSnoop_txt_IntercomService() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "intercom-service-2.23.0.tgz")}})
	s.Require().NoError(err)
	s.compactGoldenTest("intercom-service.results", results)
}

func (s *GoldenTest) TestSnoop_txt_Guardian() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "guardian-0.24.4.tgz")}})
	s.Require().NoError(err)
	s.compactGoldenTest("guardian.results", results)
}

func (s *GoldenTest) TestSnoop_json_Redis() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "redis-0.25.6.tgz")}})
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	actual := results[0].toJSON()
	s.EqualGoldenJSON("redis.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_txt_Redis() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "redis-0.25.6.tgz")}})
	s.Require().NoError(err)
	s.compactGoldenTest("redis.results", results)
}

func (s *GoldenTest) TestSnoop_json_RabbitmqClusterOperator() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "rabbitmq-cluster-operator-0.2.1.tgz")}})
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	actual := results[0].toJSON()
	s.EqualGoldenJSON("rabbitmq-cluster-operator.results.golden.json", actual)
}

func (s *GoldenTest) TestSnoop_txt_RabbitmqClusterOperator() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "rabbitmq-cluster-operator-0.2.1.tgz")}})
	s.Require().NoError(err)
	s.compactGoldenTest("rabbitmq-cluster-operator.results", results)
}

func (s *GoldenTest) TestSnoop_UnusedHaveValuesContext() {
	restore := disableStrictParsing()
	defer restore()

	results, err := Snoop([]ChartSettings{{Path: filepath.Join(s.chartsDir, "test-chart")}})
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	for _, p := range results[0].Unused {
		s.Require().NotEmpty(p.Contexts, "unused path %s should have a context", p.ID())
		s.Equal("values.yaml", p.Contexts[0].FileName,
			"unused path %s context should point to values.yaml", p.ID())
		s.Positive(p.Contexts[0].Line,
			"unused path %s should have a positive line number", p.ID())
	}
}

func disableStrictParsing() func() {
	oldStrict := assert.Strict
	assert.Strict = false

	return func() { assert.Strict = oldStrict }
}
