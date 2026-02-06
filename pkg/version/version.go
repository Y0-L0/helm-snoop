package version

import (
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
)

// Set via ldflags by GoReleaser / CI.
var (
	version    = "dev"
	commit     = "none"
	treeState  = "unknown"
	commitDate = "unknown"
)

type info struct {
	version    string
	commit     string
	treeState  string
	commitDate string
}

func resolve() info {
	i := info{version, commit, treeState, commitDate}

	// If ldflags were set (e.g. by GoReleaser), use them as-is
	if version != "dev" {
		return i
	}

	// Fallback to debug.ReadBuildInfo() for go install / go run
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return i
	}

	i.version = bi.Main.Version

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			i.commit = s.Value
		case "vcs.time":
			i.commitDate = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				i.treeState = "dirty"
			} else {
				i.treeState = "clean"
			}
		}
	}

	return i
}

// BuildInfo writes version information to the provided writer.
func BuildInfo(w io.Writer) {
	i := resolve()

	fmt.Fprintf(w, "Version:    %s\n", i.version)
	fmt.Fprintf(w, "Commit:     %s\n", i.commit)
	fmt.Fprintf(w, "TreeState:  %s\n", i.treeState)
	fmt.Fprintf(w, "CommitDate: %s\n", i.commitDate)
	fmt.Fprintf(w, "GoVersion:  %s\n", runtime.Version())
	fmt.Fprintf(w, "Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
