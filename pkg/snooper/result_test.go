package snooper

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

func (s *Unittest) TestHasFindings() {
	tests := []struct {
		name   string
		result *Result
		want   bool
	}{
		{
			"has Unused findings",
			&Result{
				Unused:    path.Paths{path.NewPath("unused")},
				Undefined: path.Paths{},
			},
			true,
		},
		{
			"has Undefined findings",
			&Result{
				Unused:    path.Paths{},
				Undefined: path.Paths{path.NewPath("undefined")},
			},
			true,
		},
		{
			"has both types of findings",
			&Result{
				Unused:    path.Paths{path.NewPath("unused")},
				Undefined: path.Paths{path.NewPath("undefined")},
			},
			true,
		},
		{
			"no findings",
			&Result{
				Unused:    path.Paths{},
				Undefined: path.Paths{},
			},
			false,
		},
		{
			"only Referenced (no findings)",
			&Result{
				Referenced: path.Paths{path.NewPath("used")},
				Unused:     path.Paths{},
				Undefined:  path.Paths{},
			},
			false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			got := tc.result.HasFindings()
			s.Equal(tc.want, got)
		})
	}
}

func (s *GoldenTest) compactGoldenTest(name string, results Results) {
	var buf bytes.Buffer
	results.ToText(&buf)

	goldenPath := filepath.Join("testdata", name+".golden")
	if s.update {
		err := os.WriteFile(goldenPath, buf.Bytes(), 0644)
		s.Require().NoError(err)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	s.Require().NoError(err)
	s.Equal(string(expected), buf.String())
}

func (s *GoldenTest) TestCompactEmpty() {
	results := Results{{
		ChartName: "test-chart",
		Unused:    path.Paths{},
		Undefined: path.Paths{},
	}}
	s.compactGoldenTest("compact_empty", results)
}

func (s *GoldenTest) TestCompactUnusedOnly() {
	unused1 := path.NewPath("config", "database", "host")
	unused1.Contexts = []path.PathContext{
		{FileName: "values.yaml", Line: 5, Column: 3},
	}
	unused2 := path.NewPath("podLabels")
	unused2.Contexts = []path.PathContext{
		{FileName: "values.yaml", Line: 12, Column: 1},
	}

	results := Results{{
		ChartName: "test-chart",
		Unused:    path.Paths{unused1, unused2},
		Undefined: path.Paths{},
	}}
	s.compactGoldenTest("compact_unused_only", results)
}

func (s *GoldenTest) TestCompactUndefinedOnly() {
	undef1 := path.NewPath("service", "nodePort")
	undef1.Contexts = []path.PathContext{
		{FileName: "templates/service.yaml", Line: 36, Column: 20},
	}

	results := Results{{
		ChartName: "test-chart",
		Unused:    path.Paths{},
		Undefined: path.Paths{undef1},
	}}
	s.compactGoldenTest("compact_undefined_only", results)
}

func (s *GoldenTest) TestCompactBothSections() {
	unused := path.NewPath("redis", "auth", "username")
	unused.Contexts = []path.PathContext{
		{FileName: "values.yaml", Line: 8, Column: 5},
	}

	undef := path.NewPath("provisioning", "extraLabels")
	undef.Contexts = []path.PathContext{
		{FileName: "templates/configmap.yaml", Line: 17, Column: 3},
	}

	results := Results{{
		ChartName: "my-chart",
		Unused:    path.Paths{unused},
		Undefined: path.Paths{undef},
	}}
	s.compactGoldenTest("compact_both", results)
}

func (s *GoldenTest) TestCompactMultipleCharts() {
	unused1 := path.NewPath("podLabels")
	unused1.Contexts = []path.PathContext{
		{FileName: "values.yaml", Line: 11, Column: 1},
	}
	undef1 := path.NewPath("service", "nodePort")
	undef1.Contexts = []path.PathContext{
		{FileName: "templates/service.yaml", Line: 36, Column: 20},
	}

	unused2 := path.NewPath("global", "domain")
	unused2.Contexts = []path.PathContext{
		{FileName: "values.yaml", Line: 2, Column: 3},
	}

	results := Results{
		{
			ChartName: "chart-a",
			Unused:    path.Paths{unused1},
			Undefined: path.Paths{undef1},
		},
		{
			ChartName: "chart-b",
			Unused:    path.Paths{unused2},
			Undefined: path.Paths{},
		},
	}
	s.compactGoldenTest("compact_multiple_charts", results)
}
