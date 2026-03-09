package snooper

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/y0-l0/helm-snoop/pkg/termcolor"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

func (cs Charts) ToText(w io.Writer) {
	for _, c := range cs {
		formatChartFindings(w, c)
	}
	formatSummaryHeader(w, cs)
	formatSummaryTable(w, cs)
	formatSummaryFooter(w, cs)
}

func formatChartFindings(w io.Writer, c *Chart) {
	if !c.hasFindings() {
		return
	}

	fmt.Fprintf(w, "%s\n\n", termcolor.Header(c.Name, "="))
	formatSection(w, "Unused", c.Result.Unused)
	formatSection(w, "Undefined", c.Result.Undefined)
	fmt.Fprintln(w)
}

func formatSection(w io.Writer, title string, paths vpath.Paths) {
	if len(paths) == 0 {
		return
	}

	fmt.Fprintln(w, termcolor.Header(title, "-"))
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

func formatSummaryHeader(w io.Writer, charts Charts) {
	if charts.HasFindings() {
		fmt.Fprintf(w, "%s\n\n", termcolor.RedHeader("Summary", "="))
	} else {
		fmt.Fprintf(w, "%s\n\n", termcolor.GreenHeader("Summary", "="))
	}
}

func formatSummaryTable(w io.Writer, charts Charts) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, c := range charts {
		switch {
		case c.Skip:
			fmt.Fprintf(tw, "%s\t%s\n", termcolor.Dim("s"), termcolor.Dim(c.Name))
		case c.hasFindings():
			fmt.Fprintf(
				tw, "%s\t%s\t%s\t%s\n",
				termcolor.Red("\u2717"),
				termcolor.Red(c.Name),
				termcolor.Red(fmt.Sprintf("%d Unused", len(c.Result.Unused))),
				termcolor.Red(fmt.Sprintf("%d Undefined", len(c.Result.Undefined))),
			)
		default:
			fmt.Fprintf(tw, "%s\t%s\n", termcolor.Green("\u2713"), termcolor.Green(c.Name))
		}
	}
	tw.Flush()
	fmt.Fprintln(w)
}

func formatSummaryFooter(w io.Writer, charts Charts) {
	skippedSuffix := ""
	if skipped := charts.skipped(); skipped > 0 {
		skippedSuffix = fmt.Sprintf(" (%d skipped)", skipped)
	}

	if charts.HasFindings() {
		footer := fmt.Sprintf(
			"%d Unused  %d Undefined across %d charts%s",
			charts.unused(),
			charts.undefined(),
			charts.scanned(),
			skippedSuffix,
		)
		fmt.Fprintln(w, termcolor.RedHeader(footer, "="))
	} else {
		footer := fmt.Sprintf("All %d charts ok%s", charts.scanned(), skippedSuffix)
		fmt.Fprintln(w, termcolor.GreenHeader(footer, "="))
	}
}
