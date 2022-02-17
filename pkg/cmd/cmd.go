package cmd

import (
	"context"
	"os"

	"github.com/bwplotka/mdox/pkg/clilog"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/observatorium/obsctl/pkg/version"
	"github.com/spf13/cobra"
)

const (
	logFormatLogfmt = "logfmt"
	logFormatJson   = "json"
	logFormatCLILog = "clilog"
)

var logger log.Logger

var logLevel, logFormat string

func setupLogger() {
	var lvl level.Option
	switch logLevel {
	case "error":
		lvl = level.AllowError()
	case "warn":
		lvl = level.AllowWarn()
	case "info":
		lvl = level.AllowInfo()
	case "debug":
		lvl = level.AllowDebug()
	default:
		panic("unexpected log level")
	}
	switch logFormat {
	case logFormatJson:
		logger = level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), lvl)
	case logFormatLogfmt:
		logger = level.NewFilter(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)), lvl)
	case logFormatCLILog:
		fallthrough
	default:
		logger = level.NewFilter(clilog.New(log.NewSyncWriter(os.Stderr)), lvl)
	}
}

var rootCmd = &cobra.Command{
	Use:     "obsctl",
	Short:   "CLI to interact with Observatorium",
	Long:    `CLI to interact with Observatorium`,
	Version: version.Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// TODO(saswatamcode): Propagte ctx here.
		setupLogger()
	},
	Run: func(cmd *cobra.Command, args []string) {},
}

func Execute(ctx context.Context) error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "info", "Log filtering level.")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log.format", logFormatCLILog, "Log format to use.")
}
