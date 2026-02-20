package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	filepath "path/filepath"

	"github.com/spf13/cobra"
	"github.com/y0-l0/helm-snoop/internal/assert"
	"github.com/y0-l0/helm-snoop/pkg/color"
	"github.com/y0-l0/helm-snoop/pkg/path"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
	"github.com/y0-l0/helm-snoop/pkg/version"
)

type CliArgumentError string

func (e CliArgumentError) Error() string { return string(e) }

// gzipMagic is the gzip file signature (RFC 1952).
var gzipMagic = []byte{0x1F, 0x8B, 0x08}

// isGzipFile checks the first bytes of a file for the gzip magic number.
func isGzipFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 3)
	n, _ := f.Read(buf)
	return n == 3 && bytes.Equal(buf, gzipMagic)
}

// resolveChartRoot finds the chart root for a given path.
// Archives are returned as-is; other paths walk up to find Chart.yaml.
func resolveChartRoot(filePath string) (string, error) {
	dir, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path for %s: %w", filePath, err)
	}

	if isGzipFile(dir) {
		return dir, nil
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "Chart.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no Chart.yaml found in any parent of %s", filePath)
		}
		dir = parent
	}
}

// resolveUniqueChartRoots resolves each path to its chart root and deduplicates.
func resolveUniqueChartRoots(paths []string) ([]string, error) {
	seen := map[string]struct{}{}
	var roots []string
	for _, p := range paths {
		root, err := resolveChartRoot(p)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[root]; !ok {
			seen[root] = struct{}{}
			roots = append(roots, root)
		}
	}
	return roots, nil
}

func analyze(
	chartPath string,
	ignorePaths *cliPaths,
	jsonOutput bool,
	showReferenced bool,
	outWriter io.Writer,
	snoop snooper.SnoopFunc,
) error {
	assert.Strict = false

	result, err := snoop(chartPath, path.Paths(*ignorePaths))
	if err != nil {
		return err
	}

	if jsonOutput {
		if err := result.ToJSON(outWriter, showReferenced); err != nil {
			return errors.New("")
		}
	} else {
		snooper.Results{result}.ToText(outWriter)
	}

	if result.HasFindings() {
		return errors.New("")
	}
	return nil
}

func NewParser(args []string, setupLogging func(slog.Level), snoop snooper.SnoopFunc) *cobra.Command {
	slog.Debug("raw cli arguments", "args", args)

	var verbosity int
	var noColor bool
	ignorePaths := &cliPaths{}
	var jsonOutput bool
	var showReferenced bool

	rootCmd := &cobra.Command{
		Use:   "helm-snoop [FLAGS] <chart-path or file>...",
		Short: "Analyze Helm charts for unused and undefined values",
		Long: `helm-snoop analyzes Helm charts to identify:
  - Values defined but never used in templates
  - Values referenced in templates but not defined in values.yaml

Examples:
  helm-snoop ./my-chart/
  helm-snoop ./my-other-chart/values.yaml`,
		Args:          cobra.MinimumNArgs(1),
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if noColor {
				color.Disable()
			}
			var logLevel slog.Level
			switch verbosity {
			case 0:
				logLevel = slog.LevelWarn
			case 1:
				logLevel = slog.LevelInfo
			default:
				logLevel = slog.LevelDebug
			}
			setupLogging(logLevel)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			chartRoots, err := resolveUniqueChartRoots(args)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			var firstErr error
			for _, root := range chartRoots {
				if err := analyze(root, ignorePaths, jsonOutput, showReferenced, cmd.OutOrStdout(), snoop); err != nil {
					firstErr = err
				}
			}
			return firstErr
		},
	}

	rootCmd.SetArgs(args[1:])

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			version.BuildInfo(cmd.OutOrStdout())
		},
	}

	rootCmd.PersistentFlags().CountVarP(
		&verbosity,
		"verbose",
		"v",
		"Increase the log level. Can be specified multiple times.",
	)

	rootCmd.PersistentFlags().BoolVar(
		&noColor,
		"no-color",
		false,
		"Disable colored output",
	)

	rootCmd.Flags().VarP(
		ignorePaths,
		"ignore",
		"i",
		`Ignore value paths matching patterns. Supports wildcards (*) and integers match as any key.
Examples:
  -i .image.tag        Ignore exact path
  -i .config.*         Ignore all config descendants
  -i .items.0          Ignore items[0] and items["0"]
  -i .a.*.c            Ignore .a.<any>.c (one level)
Repeatable. Paths match the dot-notation output format for easy copy-paste.`,
	)

	rootCmd.Flags().BoolVar(
		&jsonOutput,
		"json",
		false,
		"Output results in JSON format",
	)

	rootCmd.Flags().BoolVar(
		&showReferenced,
		"referenced",
		false,
		"Include referenced values in output",
	)

	rootCmd.AddCommand(versionCmd)
	return rootCmd
}

func Main(args []string, stdout io.Writer, stderr io.Writer, setupLogging func(slog.Level), snoop snooper.SnoopFunc) int {
	command := NewParser(args, setupLogging, snoop)

	command.SetOut(stdout)
	command.SetErr(stderr)

	err := command.Execute()
	if err != nil {
		if err.Error() != "" {
			command.PrintErrln(err.Error())
		}
		return 1
	}

	return 0
}
