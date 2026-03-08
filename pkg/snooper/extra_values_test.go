package snooper

import (
	"path/filepath"
)

func (s *GoldenTest) TestSnoop_ExtraValues_AddsDefinitions() {
	restore := disableStrictParsing()
	defer restore()

	// test-chart has no findings normally. Adding extraValues introduces
	// a new defined-but-unused path.
	charts := []ChartSettings{{
		Path: filepath.Join(s.chartsDir, "test-chart"),
		ExtraValues: map[string]any{
			"extraKey": "extraValue",
		},
	}}

	results, err := Snoop(charts)
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	var unusedIDs []string
	for _, p := range results[0].Unused {
		unusedIDs = append(unusedIDs, p.ID())
	}
	s.Contains(unusedIDs, ".extraKey", "extraValues key should appear as unused")
}

func (s *GoldenTest) TestSnoop_ExtraValues_NestedMap() {
	restore := disableStrictParsing()
	defer restore()

	charts := []ChartSettings{{
		Path: filepath.Join(s.chartsDir, "test-chart"),
		ExtraValues: map[string]any{
			"deep": map[string]any{
				"nested": "value",
			},
		},
	}}

	results, err := Snoop(charts)
	s.Require().NoError(err)
	s.Require().Len(results, 1)

	var unusedIDs []string
	for _, p := range results[0].Unused {
		unusedIDs = append(unusedIDs, p.ID())
	}
	s.Contains(unusedIDs, ".deep.nested", "nested extraValues should appear as unused")
}

func (s *GoldenTest) TestSnoop_ExtraValues_Nil() {
	restore := disableStrictParsing()
	defer restore()

	// nil ExtraValues should behave the same as before.
	baseline, err := Snoop([]ChartSettings{{
		Path: filepath.Join(s.chartsDir, "test-chart"),
	}})
	s.Require().NoError(err)

	withNil, err := Snoop([]ChartSettings{{
		Path:        filepath.Join(s.chartsDir, "test-chart"),
		ExtraValues: nil,
	}})
	s.Require().NoError(err)

	s.Len(withNil[0].Unused, len(baseline[0].Unused))
	s.Len(withNil[0].Undefined, len(baseline[0].Undefined))
}
