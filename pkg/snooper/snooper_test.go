package snooper

import (
	"github.com/y0-l0/helm-snoop/pkg/path"
	loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

// Validates that AnalysisV2 uses Helm's loader and produces expected results
// on the simple test chart.
func (s *Integrationtest) TestAnalysisV2_SimpleChart() {
	chart, err := loader.Load(s.chartPath)
	s.Require().NoError(err)

	r, err := Snoop(chart)
	s.Require().NoError(err)

	s.Require().Contains(r.Referenced, path.NewPath("config", "enabled"))
	s.Require().Contains(r.Referenced, path.NewPath("config", "message"))
	s.Require().Empty(r.DefinedNotUsed)
	s.Require().Empty(r.UsedNotDefined)
}
