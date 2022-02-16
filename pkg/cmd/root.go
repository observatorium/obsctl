package cmd

import (
	"context"
	"fmt"
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

func setupLogger() log.Logger {
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
		return level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), lvl)
	case logFormatLogfmt:
		return level.NewFilter(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr)), lvl)
	case logFormatCLILog:
		fallthrough
	default:
		return level.NewFilter(clilog.New(log.NewSyncWriter(os.Stderr)), lvl)
	}
}

var rootCmd = &cobra.Command{
	Use:     "obsctl",
	Short:   "CLI to interact with Observatorium",
	Long:    `CLI to interact with Observatorium`,
	Version: version.Version,
	Run: func(cmd *cobra.Command, args []string) {
		logger = setupLogger()
	},
}

func Execute(ctx context.Context) error {
	if err := rootCmd.Execute(); err != nil {
		level.Error(logger).Log("err", fmt.Sprintf("%+v", fmt.Errorf("command exec failed %w", err)))
		return err
	}
	// TODO(saswatamcode): Propagte ctx here.
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "info", "Log filtering level.")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log.format", logFormatCLILog, "Log format to use.")
}
