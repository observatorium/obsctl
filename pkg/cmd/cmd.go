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

var logLevel, logFormat string
var logger log.Logger

func setupLogger(*cobra.Command, []string) {
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

func NewObsctlCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "obsctl",
		Short:            "CLI to interact with Observatorium",
		Long:             `CLI to interact with Observatorium`,
		Version:          version.Version,
		PersistentPreRun: setupLogger,
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "run called")
		},
	}

	cmd.AddCommand(NewMetricsCmd(ctx))
	cmd.AddCommand(NewContextCommand(ctx))
	cmd.AddCommand(NewLoginCmd(ctx))

	cmd.PersistentFlags().StringVar(&logLevel, "log.level", "info", "Log filtering level.")
	cmd.PersistentFlags().StringVar(&logFormat, "log.format", logFormatCLILog, "Log format to use.")

	return cmd
}
