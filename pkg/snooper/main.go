package snooper

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	loader "helm.sh/helm/v4/pkg/chart/v2/loader"
)

// Simple entry point for CLI-style invocation.
// Expects exactly one positional argument: the chart path.
// Writes a plain-text report to stdout and returns an exit code.
// Exit codes: 0 success; 1 analysis error; 2 usage error.
func Main(args []string, stdout io.Writer, stderr io.Writer) int {
	// Optional second CLI arg can specify log level: debug|info|warn|error
	level := slog.LevelInfo
	if len(args) >= 3 {
		level = parseLevel(args[2])
		args = args[:2]
	}
	SetupLogging(level)
	if len(args) != 2 {
		_, _ = fmt.Fprintln(stderr, "usage: helm-snoop <path-to-chart> [debug|info|warn|error]")
		return 2
	}

	chartPath := args[1]
	chart, err := loader.Load(chartPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Failed to read helm chart.\nerror: %v\n", err)
		return 1
	}
	r, err := Snoop(chart)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Failed to analyze the helm chart.\nerror: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "Referenced:")
	for _, p := range r.Referenced {
		_, _ = fmt.Fprintln(stdout, p.ID())
	}

	_, _ = fmt.Fprintln(stdout, "Defined-not-used:")
	for _, p := range r.DefinedNotUsed {
		_, _ = fmt.Fprintln(stdout, p.ID())
	}

	_, _ = fmt.Fprintln(stdout, "Used-not-defined:")
	for _, p := range r.UsedNotDefined {
		_, _ = fmt.Fprintln(stdout, p.ID())
	}

	return 0
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
