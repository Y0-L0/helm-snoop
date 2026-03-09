package snooper

import (
	"path/filepath"
)

func (s *GoldenTest) TestSnoop_ExtraValues_AddsDefinitions() {
	restore := disableStrictParsing()
	defer restore()

	// test-chart has no findings normally. Adding extraValues introduces
	// a new defined-but-unused path.
	charts := Charts{{
		Path: filepath.Join(s.chartsDir, "test-chart"),
		ExtraValues: map[string]any{
			"extraKey": "extraValue",
		},
	}}

	s.Require().NoError(Snoop(charts))

	var unusedIDs []string
	for _, p := range charts[0].Result.Unused {
		unusedIDs = append(unusedIDs, p.ID())
	}
	s.Contains(unusedIDs, ".extraKey", "extraValues key should appear as unused")
}

func (s *GoldenTest) TestSnoop_ExtraValues_NestedMap() {
	restore := disableStrictParsing()
	defer restore()

	charts := Charts{{
		Path: filepath.Join(s.chartsDir, "test-chart"),
		ExtraValues: map[string]any{
			"deep": map[string]any{
				"nested": "value",
			},
		},
	}}

	s.Require().NoError(Snoop(charts))

	var unusedIDs []string
	for _, p := range charts[0].Result.Unused {
		unusedIDs = append(unusedIDs, p.ID())
	}
	s.Contains(unusedIDs, ".deep.nested", "nested extraValues should appear as unused")
}

func (s *GoldenTest) TestSnoop_ExtraValues_Nil() {
	restore := disableStrictParsing()
	defer restore()

	// nil ExtraValues should behave the same as before.
	baseline := Charts{{Path: filepath.Join(s.chartsDir, "test-chart")}}
	s.Require().NoError(Snoop(baseline))

	withNil := Charts{{
		Path:        filepath.Join(s.chartsDir, "test-chart"),
		ExtraValues: nil,
	}}
	s.Require().NoError(Snoop(withNil))

	s.Len(withNil[0].Result.Unused, len(baseline[0].Result.Unused))
	s.Len(withNil[0].Result.Undefined, len(baseline[0].Result.Undefined))
}
