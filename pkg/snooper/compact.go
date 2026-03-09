// Package snooper analyzes Helm charts for unused and undefined value paths.
package snooper

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/y0-l0/helm-snoop/pkg/termcolor"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

func formatChartCompact(w io.Writer, c *Chart) {
	// Chart header
	fmt.Fprintf(w, "%s\n\n", termcolor.Header(c.Name, "="))

	if c.Result == nil {
		return
	}

	// Unused section (only if non-empty)
	if len(c.Result.Unused) > 0 {
		fmt.Fprintln(w, termcolor.Header("Unused", "-"))
		formatPathsCompact(w, c.Result.Unused)
	}

	// Undefined section (only if non-empty)
	if len(c.Result.Undefined) > 0 {
		fmt.Fprintln(w, termcolor.Header("Undefined", "-"))
		formatPathsCompact(w, c.Result.Undefined)
	}

	fmt.Fprintln(w)
}

func formatSummary(w io.Writer, charts Charts) {
	fmt.Fprintf(w, "%s\n\n", termcolor.Header("Summary", "="))

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	totalUnused := 0
	totalUndefined := 0
	for _, c := range charts {
		unused := 0
		undefined := 0
		if c.Result != nil {
			unused = len(c.Result.Unused)
			undefined = len(c.Result.Undefined)
		}
		totalUnused += unused
		totalUndefined += undefined
		fmt.Fprintf(tw, "%s\t%d Unused\t%d Undefined\t\n", c.Name, unused, undefined)
	}
	fmt.Fprintf(tw, "Total\t%d Unused\t%d Undefined\tacross %d chart(s)\n", totalUnused, totalUndefined, len(charts))
	tw.Flush()
}

func formatPathsCompact(w io.Writer, paths vpath.Paths) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.StripEscape)
	for _, p := range paths {
		if len(p.Contexts) == 0 {
			fmt.Fprintf(tw, "%s\n", termcolor.Red(p.ID()))
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\n", termcolor.Red(p.ID()), termcolor.Dim(p.Contexts[0].String()))
		for _, ctx := range p.Contexts[1:] {
			fmt.Fprintf(tw, "%s\t%s\n", termcolor.Red(""), termcolor.Dim(ctx.String()))
		}
	}
	tw.Flush()
}
