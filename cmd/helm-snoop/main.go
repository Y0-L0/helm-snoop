package main

import (
	"io"
	"os"

	"github.com/y0-l0/helm-snoop/pkg/cli"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
)

// injectable process controls for testing main() without exiting the test binary
var (
	osExit           = os.Exit
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func main() {
	osExit(cli.Main(os.Args, stdout, stderr, snooper.SetupLogging, snooper.Snoop))
}
