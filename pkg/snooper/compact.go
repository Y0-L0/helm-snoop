package snooper

import (
	"fmt"
	"io"
	"strings"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

const headerWidth = 34

// FormatCompact writes the compact view output for multiple chart results.
func FormatCompact(w io.Writer, results []*Result) {
	for _, result := range results {
		result.ToText(w)
	}
}

func formatChartCompact(w io.Writer, result *Result) {
	// Chart header
	fmt.Fprintln(w, centerHeader(result.ChartName, "="))
	fmt.Fprintln(w)

	// Unused section (only if non-empty)
	if len(result.Unused) > 0 {
		fmt.Fprintln(w, centerHeader("Unused", "-"))
		formatPathsCompact(w, result.Unused)
		fmt.Fprintln(w)
	}

	// Undefined section (only if non-empty)
	if len(result.Undefined) > 0 {
		fmt.Fprintln(w, centerHeader("Undefined", "-"))
		formatPathsCompact(w, result.Undefined)
		fmt.Fprintln(w)
	}
}

func formatPathsCompact(w io.Writer, paths path.Paths) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, p := range paths {
		location := ""
		if len(p.Contexts) > 0 {
			location = p.Contexts[0].String()
		}
		fmt.Fprintf(tw, "%s\t%s\n", p.ID(), location)
	}
	tw.Flush()
}

// centerHeader creates a centered header like "=========== name ==========="
func centerHeader(text string, char string) string {
	textWithSpaces := " " + text + " "
	totalPadding := headerWidth - len(textWithSpaces)
	if totalPadding < 2 {
		return char + textWithSpaces + char
	}
	leftPad := totalPadding / 2
	rightPad := totalPadding - leftPad
	return strings.Repeat(char, leftPad) + textWithSpaces + strings.Repeat(char, rightPad)
}
