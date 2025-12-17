package snooper

import (
	"fmt"
	"io"

	loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

// Simple entry point for CLI-style invocation.
// Expects exactly one positional argument: the chart path.
// Writes a plain-text report to stdout and returns an exit code.
// Exit codes: 0 success; 1 analysis error; 2 usage error.
func Main(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 2 {
		_, _ = fmt.Fprintln(stderr, "usage: helm-snoop <path-to-chart>")
		return 2
	}

	chartPath := args[1]
	chart, err := loader.Load(chartPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Failed to read helm chart.\nerror: %v\n", err)
		return 1
	}
	r, err := Analyse(chart)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Failed to analyze the helm chart.\nerror: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintln(stdout, "Referenced:")
	for _, k := range r.Referenced {
		_, _ = fmt.Fprintf(stdout, "  - %s\n", k)
	}

	_, _ = fmt.Fprintln(stdout, "Defined-not-used:")
	for _, k := range r.DefinedNotUsed {
		_, _ = fmt.Fprintf(stdout, "  - %s\n", k)
	}

	_, _ = fmt.Fprintln(stdout, "Used-not-defined:")
	for _, k := range r.UsedNotDefined {
		_, _ = fmt.Fprintf(stdout, "  - %s\n", k)
	}

	return 0
}
