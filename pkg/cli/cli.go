package cli

import (
	"errors"
	"io"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/y0-l0/helm-snoop/pkg/parser"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
	"github.com/y0-l0/helm-snoop/pkg/version"
)

type CliArgumentError string

func (e CliArgumentError) Error() string { return string(e) }

func analyze(
	chartPath string,
	ignoreKeys []string,
	jsonOutput bool,
	outWriter io.Writer,
	snoop snooper.SnoopFunc,
) error {
	parser.Strict = false

	result, err := snoop(chartPath, ignoreKeys)
	if err != nil {
		return err
	}

	if jsonOutput {
		if err := result.ToJSON(outWriter); err != nil {
			return errors.New("")
		}
	} else {
		if err := result.ToText(outWriter); err != nil {
			return errors.New("")
		}
	}

	if result.HasFindings() {
		return errors.New("")
	}
	return nil
}

func NewParser(args []string, setupLogging func(slog.Level), snoop snooper.SnoopFunc) *cobra.Command {
	slog.Debug("raw cli arguments", "args", args)

	var verbosity int
	var ignoreKeys []string
	var jsonOutput bool

	rootCmd := &cobra.Command{
		Use:   "helm-snoop [FLAGS] <chart-path>",
		Short: "Analyze Helm charts for unused and undefined values",
		Long: `helm-snoop analyzes Helm charts to identify:
  - Values defined but never used in templates
  - Values referenced in templates but not defined in values.yaml`,
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
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
			return analyze(args[0], ignoreKeys, jsonOutput, cmd.OutOrStdout(), snoop)
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

	rootCmd.Flags().StringArrayVarP(
		&ignoreKeys,
		"ignore",
		"i",
		nil,
		"Ignore specific keys (repeatable)",
	)

	rootCmd.Flags().BoolVar(
		&jsonOutput,
		"json",
		false,
		"Output results in JSON format",
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
