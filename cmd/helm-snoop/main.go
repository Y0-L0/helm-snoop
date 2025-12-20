package main

import (
	"io"
	"os"

	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

// injectable process controls for testing main() without exiting the test binary
var (
	osExit           = os.Exit
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func main() {
	// Disable strict panics for production runs; tests keep the default (true).
	parser.Strict = false
	osExit(snooper.Main(os.Args, stdout, stderr))
}
