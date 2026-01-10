package cli

import (
	"io"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/y0-l0/helm-snoop/pkg/snooper"
	"github.com/y0-l0/helm-snoop/pkg/version"
)

type CliArgumentError string

func (e CliArgumentError) Error() string { return string(e) }

func NewParser(args []string, setupLogging func(slog.Level), snoop snooper.SnoopFunc) *cobra.Command {
	slog.Debug("raw cli arguments", "args", args)

	config := &cliConfig{
		snoop: snoop,
	}
	var verbosity int

	rootCmd := &cobra.Command{
		Use:   "helm-snoop [FLAGS] <chart-path>",
		Short: "Analyze Helm charts for unused and undefined values",
		Long: `helm-snoop analyzes Helm charts to identify:
  - Values defined but never used in templates
  - Values referenced in templates but not defined in values.yaml`,
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
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
			config.chartPath = args[0]
			config.outWriter = cmd.OutOrStdout()
			slog.Debug("Parsed cli arguments", "config", config)
			return config.analyze()
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
		&config.ignoreKeys,
		"ignore",
		"i",
		nil,
		"Ignore specific keys (repeatable)",
	)

	rootCmd.Flags().BoolVar(
		&config.jsonOutput,
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
			command.PrintErrln("Error:", err.Error())
		}
		return 1
	}

	return 0
}
