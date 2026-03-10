// Package cli defines the helm-snoop command-line interface.
package cli

import (
	"errors"
	"io"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/y0-l0/helm-snoop/internal/assert"
	"github.com/y0-l0/helm-snoop/pkg/appinfo"
	"github.com/y0-l0/helm-snoop/pkg/config"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
	"github.com/y0-l0/helm-snoop/pkg/termcolor"
	"github.com/y0-l0/helm-snoop/pkg/vpath"
)

type ArgumentError string

func (e ArgumentError) Error() string { return string(e) }

func NewParser(args []string, setupLogging func(slog.Level), snoop snooper.SnoopFunc) *cobra.Command {
	slog.Debug("raw cli arguments", "args", args)

	var verbosity int
	var noColor bool
	ignorePaths := &cliPaths{}
	var valuesFiles []string
	var configPath string
	var noConfig bool
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
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if noColor {
				termcolor.Disable()
			}
			var logLevel slog.Level
			switch verbosity {
			case 0:
				logLevel = slog.LevelError
			case 1:
				logLevel = slog.LevelWarn
			case 2:
				logLevel = slog.LevelInfo
			default:
				logLevel = slog.LevelDebug
			}
			setupLogging(logLevel)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			chartRoots, err := resolveUniqueCharts(args)
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			assert.Strict = false //nolint:reassign // strict mode is for dev/test only.

			charts, err := config.Resolve(chartRoots, config.Options{
				ConfigPath:  configPath,
				NoConfig:    noConfig,
				Ignore:      vpath.Paths(*ignorePaths),
				ValuesFiles: valuesFiles,
			})
			if err != nil {
				return err
			}

			if err := snoop(charts); err != nil {
				return err
			}

			if !jsonOutput {
				charts.ToText(cmd.OutOrStdout())
			} else if err := charts.ToJSON(cmd.OutOrStdout(), showReferenced); err != nil {
				return errors.New("")
			}

			if charts.HasFindings() {
				return errors.New("")
			}
			return nil
		},
	}

	rootCmd.SetArgs(args[1:])

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, _ []string) {
			appinfo.Write(cmd.OutOrStdout())
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

	rootCmd.Flags().StringArrayVarP(
		&valuesFiles,
		"values",
		"f",
		nil,
		"Additional values files to include in the analysis. Repeatable.",
	)

	rootCmd.Flags().StringVar(
		&configPath,
		"config",
		"",
		"Path to .helm-snoop.yaml config file (default: auto-discover)",
	)

	rootCmd.Flags().BoolVar(
		&noConfig,
		"no-config",
		false,
		"Skip config file loading",
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

func Main(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	setupLogging func(slog.Level),
	snoop snooper.SnoopFunc,
) int {
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
