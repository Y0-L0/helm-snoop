package snooper

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

func (s *GoldenTest) compactGoldenTest(name string, charts Charts) {
	var buf bytes.Buffer
	charts.ToText(&buf)

	goldenPath := filepath.Join("testdata", name+".golden")
	if s.update {
		err := os.WriteFile(goldenPath, buf.Bytes(), 0o644) //nolint:gosec // test golden file update
		s.Require().NoError(err)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	s.Require().NoError(err)
	s.Equal(string(expected), buf.String())
}

func (s *GoldenTest) TestCompactEmpty() {
	charts := Charts{{
		Name: "test-chart",
		Result: &Result{
			Unused:    vpath.Paths{},
			Undefined: vpath.Paths{},
		},
	}}
	s.compactGoldenTest("compact_empty", charts)
}

func (s *GoldenTest) TestCompactUnusedOnly() {
	unused1 := vpath.NewPath("config", "database", "host")
	unused1.Contexts = vpath.Contexts{
		{FileName: "values.yaml", Line: 5, Column: 3},
	}
	unused2 := vpath.NewPath("podLabels")
	unused2.Contexts = vpath.Contexts{
		{FileName: "values.yaml", Line: 12, Column: 1},
	}

	charts := Charts{{
		Name: "test-chart",
		Result: &Result{
			Unused:    vpath.Paths{unused1, unused2},
			Undefined: vpath.Paths{},
		},
	}}
	s.compactGoldenTest("compact_unused_only", charts)
}

func (s *GoldenTest) TestCompactUndefinedOnly() {
	undef1 := vpath.NewPath("service", "nodePort")
	undef1.Contexts = vpath.Contexts{
		{FileName: "templates/service.yaml", Line: 36, Column: 20},
	}

	charts := Charts{{
		Name: "test-chart",
		Result: &Result{
			Unused:    vpath.Paths{},
			Undefined: vpath.Paths{undef1},
		},
	}}
	s.compactGoldenTest("compact_undefined_only", charts)
}

func (s *GoldenTest) TestCompactBothSections() {
	unused := vpath.NewPath("redis", "auth", "username")
	unused.Contexts = vpath.Contexts{
		{FileName: "values.yaml", Line: 8, Column: 5},
	}

	undef := vpath.NewPath("provisioning", "extraLabels")
	undef.Contexts = vpath.Contexts{
		{FileName: "templates/configmap.yaml", Line: 17, Column: 3},
	}

	charts := Charts{{
		Name: "my-chart",
		Result: &Result{
			Unused:    vpath.Paths{unused},
			Undefined: vpath.Paths{undef},
		},
	}}
	s.compactGoldenTest("compact_both", charts)
}

func (s *GoldenTest) TestCompactMultipleContexts() {
	undef := vpath.NewPath("service", "nodePort")
	undef.Contexts = vpath.Contexts{
		{FileName: "templates/service.yaml", Line: 36, Column: 20},
		{FileName: "templates/deployment.yaml", Line: 12, Column: 8},
		{FileName: "templates/ingress.yaml", Line: 5, Column: 15},
	}

	charts := Charts{{
		Name: "test-chart",
		Result: &Result{
			Unused:    vpath.Paths{},
			Undefined: vpath.Paths{undef},
		},
	}}
	s.compactGoldenTest("compact_multiple_contexts", charts)
}

func (s *GoldenTest) TestCompactMultipleCharts() {
	unused1 := vpath.NewPath("podLabels")
	unused1.Contexts = vpath.Contexts{
		{FileName: "values.yaml", Line: 11, Column: 1},
	}
	undef1 := vpath.NewPath("service", "nodePort")
	undef1.Contexts = vpath.Contexts{
		{FileName: "templates/service.yaml", Line: 36, Column: 20},
	}

	unused2 := vpath.NewPath("global", "domain")
	unused2.Contexts = vpath.Contexts{
		{FileName: "values.yaml", Line: 2, Column: 3},
	}

	charts := Charts{
		{
			Name: "chart-a",
			Result: &Result{
				Unused:    vpath.Paths{unused1},
				Undefined: vpath.Paths{undef1},
			},
		},
		{
			Name: "chart-b",
			Result: &Result{
				Unused:    vpath.Paths{unused2},
				Undefined: vpath.Paths{},
			},
		},
	}
	s.compactGoldenTest("compact_multiple_charts", charts)
}
