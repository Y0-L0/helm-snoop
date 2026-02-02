package snooper

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/y0-l0/helm-snoop/pkg/path"
)

const headerWidth = 34

func formatChartCompact(w io.Writer, result *Result) {
	// Chart header
	fmt.Fprintln(w, centerHeader(result.ChartName, "="))
	fmt.Fprintln(w)

	// Unused section (only if non-empty)
	if len(result.Unused) > 0 {
		fmt.Fprintln(w, centerHeader("Unused", "-"))
		formatPathsCompact(w, result.Unused)
	}

	// Undefined section (only if non-empty)
	if len(result.Undefined) > 0 {
		fmt.Fprintln(w, centerHeader("Undefined", "-"))
		formatPathsCompact(w, result.Undefined)
	}

	fmt.Fprintln(w)
}

func formatSummary(w io.Writer, results Results) {
	fmt.Fprintln(w, centerHeader("Summary", "="))
	fmt.Fprintln(w)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	totalUnused := 0
	totalUndefined := 0
	for _, r := range results {
		totalUnused += len(r.Unused)
		totalUndefined += len(r.Undefined)
		fmt.Fprintf(tw, "%s\t%d Unused\t%d Undefined\t\n", r.ChartName, len(r.Unused), len(r.Undefined))
	}
	fmt.Fprintf(tw, "Total\t%d Unused\t%d Undefined\tacross %d chart(s)\n", totalUnused, totalUndefined, len(results))
	tw.Flush()
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
