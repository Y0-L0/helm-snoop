package appinfo

import (
	"bytes"
	"runtime/debug"
)

// setVars overrides the package-level ldflags vars and returns a restore function.
func setVars(_version, _commit, _treeState, _commitDate string) func() {
	origVersion, origCommit, origTreeState, origCommitDate := version, commit, treeState, commitDate
	version, commit, treeState, commitDate = _version, _commit, _treeState, _commitDate
	return func() {
		version, commit, treeState, commitDate = origVersion, origCommit, origTreeState, origCommitDate
	}
}

func (s *Unittest) TestWrite() {
	var buf bytes.Buffer
	Write(&buf)

	output := buf.String()

	expectedFields := []string{"Version:", "Commit:", "TreeState:", "CommitDate:", "GoVersion:", "Platform:"}
	for _, field := range expectedFields {
		s.Contains(output, field)
	}

	// When running via go test, ReadBuildInfo sets Main.Version to "(devel)".
	s.Contains(output, "Version:    (devel)")
}

func (s *Unittest) TestResolve_Ldflags() {
	restore := setVars("1.2.3", "abc123", "clean", "2025-01-01")
	defer restore()

	i := resolve(nil)

	s.Equal("1.2.3", i.version)
	s.Equal("abc123", i.commit)
	s.Equal("clean", i.treeState)
	s.Equal("2025-01-01", i.commitDate)
}

func (s *Unittest) TestResolve_BuildInfoNotAvailable() {
	restore := setVars("dev", "none", "unknown", "unknown")
	defer restore()

	i := resolve(func() (*debug.BuildInfo, bool) { return nil, false })

	s.Equal("dev", i.version)
	s.Equal("none", i.commit)
	s.Equal("unknown", i.treeState)
	s.Equal("unknown", i.commitDate)
}

func (s *Unittest) TestResolve_BuildInfoClean() {
	i := resolve(func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v1.0.0"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc123"},
				{Key: "vcs.time", Value: "2025-06-01T12:00:00Z"},
				{Key: "vcs.modified", Value: "false"},
			},
		}, true
	})

	s.Equal("v1.0.0", i.version)
	s.Equal("abc123", i.commit)
	s.Equal("2025-06-01T12:00:00Z", i.commitDate)
	s.Equal("clean", i.treeState)
}

func (s *Unittest) TestResolve_BuildInfoDirty() {
	i := resolve(func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{
			Main: debug.Module{Version: "v1.0.0"},
			Settings: []debug.BuildSetting{
				{Key: "vcs.modified", Value: "true"},
			},
		}, true
	})

	s.Equal("dirty", i.treeState)
}
