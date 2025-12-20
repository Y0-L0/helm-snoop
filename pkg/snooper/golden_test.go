package snooper

import (
	"encoding/json"

	loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

func (s *GoldenTest) TestSnoop_TestChart() {
	chart, err := loader.Load(s.chartPath)
	s.Require().NoError(err)

	results, err := Snoop(chart)
	s.Require().NoError(err)

	actual := results.ToJSON()

	goldenPath := s.goldenPath("test-chart.results.golden.json")
	if s.update {
		actualData, err := json.MarshalIndent(actual, "", "  ")
		s.Require().NoError(err)
		s.writeFile(goldenPath, actualData)
	}

	var expected ResultsJSON
	s.Require().NoError(json.Unmarshal(s.readFile(goldenPath), &expected))

	s.Require().Equal(expected, actual)
}
