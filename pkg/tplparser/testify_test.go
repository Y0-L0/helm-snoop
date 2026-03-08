package tplparser

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"

	"github.com/y0-l0/helm-snoop/internal/testsuite"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

type Unittest struct {
	testsuite.LoggingSuite
}

func TestUnit(t *testing.T) { suite.Run(t, new(Unittest)) }

func (s *Unittest) parse(tmpl string) vpath.Paths {
	paths, err := parseFile("", "test.yaml", []byte(tmpl), nil)
	s.Require().NoError(err)
	return paths
}

type testFile struct{ name, data string }

func buildChart(files ...testFile) *chart.Chart {
	c := &chart.Chart{Templates: make([]*common.File, 0, len(files))}
	for _, f := range files {
		c.Templates = append(c.Templates, &common.File{Name: f.name, Data: []byte(f.data)})
	}
	return c
}

// parseChart builds a chart, creates a template index, and parses the last file.
func (s *Unittest) parseChart(files ...testFile) vpath.Paths {
	c := buildChart(files...)
	idx, err := BuildTemplateIndex(c)
	s.Require().NoError(err)
	last := c.Templates[len(c.Templates)-1]
	paths, err := parseFile("", last.Name, last.Data, idx)
	s.Require().NoError(err)
	return paths
}
