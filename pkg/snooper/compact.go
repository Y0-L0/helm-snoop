package snooper

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/y0-l0/helm-snoop/pkg/color"
	"github.com/y0-l0/helm-snoop/pkg/path"
)

func formatChartCompact(w io.Writer, result *Result) {
	// Chart header
	fmt.Fprintln(w, color.Header(result.ChartName, "="))
	fmt.Fprintln(w)

	// Unused section (only if non-empty)
	if len(result.Unused) > 0 {
		fmt.Fprintln(w, color.Header("Unused", "-"))
		formatPathsCompact(w, result.Unused)
	}

	// Undefined section (only if non-empty)
	if len(result.Undefined) > 0 {
		fmt.Fprintln(w, color.Header("Undefined", "-"))
		formatPathsCompact(w, result.Undefined)
	}

	fmt.Fprintln(w)
}

func formatSummary(w io.Writer, results Results) {
	fmt.Fprintln(w, color.Header("Summary", "="))
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
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.StripEscape)
	for _, p := range paths {
		if len(p.Contexts) == 0 {
			fmt.Fprintf(tw, "%s\n", color.Red(p.ID()))
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\n", color.Red(p.ID()), color.Dim(p.Contexts[0].String()))
		for _, ctx := range p.Contexts[1:] {
			fmt.Fprintf(tw, "%s\t%s\n", color.Red(""), color.Dim(ctx.String()))
		}
	}
	tw.Flush()
}
