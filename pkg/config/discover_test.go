package config

import (
	"os"
	"path/filepath"
)

func (s *Unittest) TestDiscover_InCurrentDir() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, configFileName)
	s.Require().NoError(os.WriteFile(configPath, []byte("version: 0"), 0o600))

	found, err := discover(dir)
	s.Require().NoError(err)
	s.Equal(configPath, found)
}

func (s *Unittest) TestDiscover_InParentDir() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, configFileName)
	s.Require().NoError(os.WriteFile(configPath, []byte("version: 0"), 0o600))

	child := filepath.Join(dir, "sub", "deep")
	s.Require().NoError(os.MkdirAll(child, 0o750))

	found, err := discover(child)
	s.Require().NoError(err)
	s.Equal(configPath, found)
}

func (s *Unittest) TestDiscover_NotFound() {
	dir := s.T().TempDir()
	found, err := discover(dir)
	s.Require().NoError(err)
	s.Empty(found)
}

func (s *Unittest) TestResolve_AutoDiscover() {
	dir := s.T().TempDir()
	configPath := filepath.Join(dir, configFileName)
	s.Require().NoError(os.WriteFile(configPath, []byte(`
version: 0
global:
  ignore:
    - .discovered
`), 0o600))

	// Change to the dir so discover finds it.
	origDir, err := os.Getwd()
	s.Require().NoError(err)
	s.Require().NoError(os.Chdir(dir))
	defer func() { _ = os.Chdir(origDir) }()

	charts, err := Resolve(
		[]string{filepath.Join(dir, "chart")},
		Options{},
	)
	s.Require().NoError(err)
	s.Require().Len(charts, 1)

	expectedIgnore := []string{".discovered"}
	var ids []string
	for _, p := range charts[0].Ignore {
		ids = append(ids, p.ID())
	}
	s.Equal(expectedIgnore, ids)
}
