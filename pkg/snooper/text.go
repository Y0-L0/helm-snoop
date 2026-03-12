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
		id := termcolor.Red(p.ID())
		for i := range max(len(p.Contexts), 1) {
			context := ""
			if i < len(p.Contexts) {
				context = termcolor.Dim(p.Contexts[i].String())
			}
			fmt.Fprintf(tw, "%s\t%s\n", id, context)
			id = termcolor.Red("")
		}
	}
	tw.Flush()
}

func formatSummaryHeader(w io.Writer, charts Charts) {
	header := termcolor.GreenHeader("Summary", "=")
	if charts.HasFindings() {
		header = termcolor.RedHeader("Summary", "=")
	}
	fmt.Fprintf(w, "%s\n\n", header)
}

func formatSummaryTable(w io.Writer, charts Charts) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, c := range charts {
		icon := termcolor.Green("\u2713")
		name := termcolor.Green(c.Name)
		var suffix string

		if c.Skip {
			icon = termcolor.Dim("s")
			name = termcolor.Dim(c.Name)
		} else if c.hasFindings() {
			icon = termcolor.Red("\u2717")
			name = termcolor.Red(c.Name)
			suffix = fmt.Sprintf(
				"\t%s\t%s",
				termcolor.Red(fmt.Sprintf("%d Unused", len(c.Result.Unused))),
				termcolor.Red(fmt.Sprintf("%d Undefined", len(c.Result.Undefined))),
			)
		}
		fmt.Fprintf(tw, "%s\t%s%s\n", icon, name, suffix)
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
