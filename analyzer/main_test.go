package analyzer

import (
	"bytes"
	"strings"
)

func (s *Integrationtest) TestMain_UsageError() {
	var out, err bytes.Buffer
	code := Main([]string{"analyzer"}, &out, &err)
	s.Require().Equal(2, code)
	s.Require().Contains(err.String(), "usage:")
}

func (s *Integrationtest) TestMain_NonexistentChart() {
	var out, err bytes.Buffer
	code := Main([]string{"analyzer", "does-not-exist"}, &out, &err)
	s.Require().Equal(1, code)
	s.Require().Contains(err.String(), "error:")
}

func (s *Integrationtest) TestMain_SimpleChart() {
	var out, err bytes.Buffer
	code := Main([]string{"analyzer", s.chartPath}, &out, &err)
	s.Require().Equal(0, code, err.String())

	s.Require().Contains(out.String(), "Referenced:")
	s.Require().True(strings.Contains(out.String(), "config.enabled") && strings.Contains(out.String(), "config.message"))
	s.Require().Contains(out.String(), "Defined-not-used:")
	s.Require().Contains(out.String(), "Used-not-defined:")
	s.Require().Equal("", err.String())
}
