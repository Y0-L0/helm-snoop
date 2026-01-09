package version

import (
	"fmt"
	"io"
	"runtime"
)

var (
	Version   = "dev"     // Default for local builds
	Commit    = "none"    // Injected: git commit SHA
	TreeState = "unknown" // Injected: clean or dirty
	BuildDate = "unknown" // Injected: RFC3339 timestamp
)

// BuildInfo writes version information to the provided writer
func BuildInfo(w io.Writer) {
	fmt.Fprintf(w, "Version:    %s\n", Version)
	fmt.Fprintf(w, "Commit:     %s\n", Commit)
	fmt.Fprintf(w, "TreeState:  %s\n", TreeState)
	fmt.Fprintf(w, "BuildDate:  %s\n", BuildDate)
	fmt.Fprintf(w, "GoVersion:  %s\n", runtime.Version())
	fmt.Fprintf(w, "Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
