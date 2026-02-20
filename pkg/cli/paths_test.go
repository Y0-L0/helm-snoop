package cli

import (
	filepath "path/filepath"

	"github.com/y0-l0/helm-snoop/pkg/path"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

func np() *path.Path { return &path.Path{} }

func (s *Unittest) TestIgnorePaths_SinglePath() {
	var capturedPaths path.Paths

	mockSnoop := func(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*snooper.Result, error) {
		capturedPaths = ignorePaths
		return &snooper.Result{
			Referenced: path.Paths{},
			Unused:     path.Paths{},
			Undefined:  path.Paths{},
		}, nil
	}

	command := NewParser(
		[]string{"helm-snoop", "-i", ".image.tag", "../../testdata/test-chart"},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)

	expected := path.Paths{np().Key("image").Key("tag")}
	path.EqualPaths(s, expected, capturedPaths)
}

func (s *Unittest) TestIgnorePaths_MultipleWithAllKinds() {
	var capturedPaths path.Paths

	mockSnoop := func(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*snooper.Result, error) {
		capturedPaths = ignorePaths
		return &snooper.Result{
			Referenced: path.Paths{},
			Unused:     path.Paths{},
			Undefined:  path.Paths{},
		}, nil
	}

	command := NewParser(
		[]string{
			"helm-snoop",
			"-i", ".image.tag", // key kind
			"-i", ".items.0", // any kind (integer)
			"-i", ".config.*", // wildcard kind (terminal)
			"-i", ".a.*.c", // wildcard kind (interior)
			"../../testdata/test-chart",
		},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)

	expected := path.Paths{
		np().Key("image").Key("tag"),
		np().Key("items").Any("0"),
		np().Key("config").Wildcard(),
		np().Key("a").Wildcard().Key("c"),
	}
	path.EqualPaths(s, expected, capturedPaths)
}

func (s *Unittest) TestIgnorePaths_InvalidPaths() {
	tests := []struct {
		name        string
		path        string
		errContains string
	}{
		{
			name:        "slash notation rejected",
			path:        "/image/tag",
			errContains: "dot notation",
		},
		{
			name:        "trailing dot",
			path:        ".config.",
			errContains: "trailing dot",
		},
		{
			name:        "empty path",
			path:        "",
			errContains: "empty pattern",
		},
		{
			name:        "double dot",
			path:        ".config..value",
			errContains: "empty segment",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			mockSnoop := func(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*snooper.Result, error) {
				s.T().Fatal("snoop should not be called with invalid path")
				return nil, nil
			}

			command := NewParser(
				[]string{"helm-snoop", "-i", tc.path, "../../testdata/test-chart"},
				snooper.SetupLogging,
				mockSnoop,
			)

			err := command.Execute()
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.errContains)
		})
	}
}

func (s *Unittest) TestIgnorePaths_NoIgnoreList() {
	var capturedPaths path.Paths

	mockSnoop := func(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*snooper.Result, error) {
		capturedPaths = ignorePaths
		return &snooper.Result{
			Referenced: path.Paths{},
			Unused:     path.Paths{},
			Undefined:  path.Paths{},
		}, nil
	}

	command := NewParser(
		[]string{"helm-snoop", "../../testdata/test-chart"},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)

	path.EqualPaths(s, path.Paths{}, capturedPaths)
}

func (s *Unittest) TestValuesFiles_SingleFile() {
	var capturedFiles []string

	mockSnoop := func(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*snooper.Result, error) {
		capturedFiles = valuesFiles
		return &snooper.Result{
			Referenced: path.Paths{},
			Unused:     path.Paths{},
			Undefined:  path.Paths{},
		}, nil
	}

	extraValues := filepath.Join(testdataDir(), "test-chart", "values.yaml")
	command := NewParser(
		[]string{"helm-snoop", "-f", extraValues, "../../testdata/test-chart"},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)
	s.Require().Equal([]string{extraValues}, capturedFiles)
}

func (s *Unittest) TestValuesFiles_MultipleFiles() {
	var capturedFiles []string

	mockSnoop := func(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*snooper.Result, error) {
		capturedFiles = valuesFiles
		return &snooper.Result{
			Referenced: path.Paths{},
			Unused:     path.Paths{},
			Undefined:  path.Paths{},
		}, nil
	}

	f1 := filepath.Join(testdataDir(), "test-chart", "values.yaml")
	f2 := filepath.Join(testdataDir(), "test-chart", "values.yaml") // reuse same file for simplicity
	command := NewParser(
		[]string{"helm-snoop", "-f", f1, "-f", f2, "../../testdata/test-chart"},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)
	s.Require().Equal([]string{f1, f2}, capturedFiles)
}

func (s *Unittest) TestValuesFiles_NoFlag() {
	var capturedFiles []string

	mockSnoop := func(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*snooper.Result, error) {
		capturedFiles = valuesFiles
		return &snooper.Result{
			Referenced: path.Paths{},
			Unused:     path.Paths{},
			Undefined:  path.Paths{},
		}, nil
	}

	command := NewParser(
		[]string{"helm-snoop", "../../testdata/test-chart"},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().NoError(err)
	s.Require().Empty(capturedFiles)
}

func (s *Unittest) TestValuesFiles_MissingFile() {
	missing := "/nonexistent/overlay-values.yaml"

	mockSnoop := func(chartPath string, ignorePaths path.Paths, valuesFiles []string) (*snooper.Result, error) {
		return snooper.Snoop(chartPath, ignorePaths, valuesFiles)
	}

	command := NewParser(
		[]string{"helm-snoop", "-f", missing, "../../testdata/test-chart"},
		snooper.SetupLogging,
		mockSnoop,
	)

	err := command.Execute()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), missing)
}
