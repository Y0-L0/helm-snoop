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
	fmt.Fprintf(w, "%s\n\n", color.Header(result.ChartName, "=")) //nolint:errcheck

	// Unused section (only if non-empty)
	if len(result.Unused) > 0 {
		fmt.Fprintln(w, color.Header("Unused", "-")) //nolint:errcheck
		formatPathsCompact(w, result.Unused)
	}

	// Undefined section (only if non-empty)
	if len(result.Undefined) > 0 {
		fmt.Fprintln(w, color.Header("Undefined", "-")) //nolint:errcheck
		formatPathsCompact(w, result.Undefined)
	}

	fmt.Fprintln(w) //nolint:errcheck
}

func formatSummary(w io.Writer, results Results) {
	fmt.Fprintf(w, "%s\n\n", color.Header("Summary", "=")) //nolint:errcheck

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	totalUnused := 0
	totalUndefined := 0
	for _, r := range results {
		totalUnused += len(r.Unused)
		totalUndefined += len(r.Undefined)
		fmt.Fprintf(tw, "%s\t%d Unused\t%d Undefined\t\n", r.ChartName, len(r.Unused), len(r.Undefined)) //nolint:errcheck
	}
	fmt.Fprintf(tw, "Total\t%d Unused\t%d Undefined\tacross %d chart(s)\n", totalUnused, totalUndefined, len(results)) //nolint:errcheck
	tw.Flush()                                                                                                         //nolint:errcheck
}

func formatPathsCompact(w io.Writer, paths path.Paths) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.StripEscape)
	for _, p := range paths {
		if len(p.Contexts) == 0 {
			fmt.Fprintf(tw, "%s\n", color.Red(p.ID())) //nolint:errcheck
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\n", color.Red(p.ID()), color.Dim(p.Contexts[0].String())) //nolint:errcheck
		for _, ctx := range p.Contexts[1:] {
			fmt.Fprintf(tw, "%s\t%s\n", color.Red(""), color.Dim(ctx.String())) //nolint:errcheck
		}
	}
	tw.Flush() //nolint:errcheck
}
