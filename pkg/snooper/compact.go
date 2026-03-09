// Package snooper analyzes Helm charts for unused and undefined value paths.
package snooper

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/y0-l0/helm-snoop/pkg/termcolor"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

func formatChartFindings(w io.Writer, c *Chart) {
	fmt.Fprintf(w, "%s\n\n", termcolor.Header(c.Name, "="))

	if len(c.Result.Unused) > 0 {
		fmt.Fprintln(w, termcolor.Header("Unused", "-"))
		formatPathsCompact(w, c.Result.Unused)
	}

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
	scanned := 0
	skipped := 0
	for _, c := range charts {
		if c.Skip {
			skipped++
			fmt.Fprintf(tw, "%s\t%s\n", termcolor.Dim("s"), termcolor.Dim(c.Name))
			continue
		}
		scanned++
		if c.Result != nil {
			totalUnused += len(c.Result.Unused)
			totalUndefined += len(c.Result.Undefined)
		}
		if c.HasFindings() {
			fmt.Fprintf(
				tw, "%s\t%s\t%d Unused\t%d Undefined\n",
				termcolor.Red("\u2717"),
				termcolor.Red(c.Name),
				len(c.Result.Unused),
				len(c.Result.Undefined),
			)
		} else {
			fmt.Fprintf(tw, "%s\t%s\n", termcolor.Green("\u2713"), termcolor.Green(c.Name))
		}
	}
	tw.Flush()

	fmt.Fprintln(w)

	hasFindings := totalUnused > 0 || totalUndefined > 0
	skippedSuffix := ""
	if skipped > 0 {
		skippedSuffix = fmt.Sprintf(" (%d skipped)", skipped)
	}

	if hasFindings {
		footer := fmt.Sprintf(
			"%d Unused  %d Undefined across %d charts%s",
			totalUnused,
			totalUndefined,
			scanned,
			skippedSuffix,
		)
		fmt.Fprintln(w, termcolor.RedHeader(footer, "="))
	} else {
		footer := fmt.Sprintf("All %d charts ok%s", scanned, skippedSuffix)
		fmt.Fprintln(w, termcolor.GreenHeader(footer, "="))
	}
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
