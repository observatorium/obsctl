package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/bwplotka/mdox/pkg/clilog"
	extflag "github.com/efficientgo/tools/extkingpin"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/observatorium/obsctl/pkg/extkingpin"
	"github.com/observatorium/obsctl/pkg/version"
	"github.com/oklog/run"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	logFormatLogfmt = "logfmt"
	logFormatJson   = "json"
	logFormatCLILog = "clilog"
)

func setupLogger(logLevel, logFormat string) log.Logger {
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

func main() {
	app := extkingpin.NewApp(kingpin.New(filepath.Base(os.Args[0]), `obsctl`).Version(version.Version))
	logLevel := app.Flag("log.level", "Log filtering level.").Default("info").Enum("error", "warn", "info", "debug")
	logFormat := app.Flag("log.format", "Log format to use.").Default(logFormatCLILog).Enum(logFormatLogfmt, logFormatJson, logFormatCLILog)

	ctx, cancel := context.WithCancel(context.Background())
	// Auth commands.
	registerLogin(ctx, app)
	registerLogout(ctx, app)
	registerCurrent(ctx, app)
	registerSwitch(ctx, app)

	// Metrics operations.
	// TODO(saswatamcode): Scope these under metrics flag, to support logs and traces in future.
	registerRead(ctx, app)
	registerRules(ctx, app)
	registerQuery(ctx, app)

	cmd, runner := app.Parse()
	logger := setupLogger(*logLevel, *logFormat)

	var g run.Group
	g.Add(func() error {
		return runner(ctx, logger)
	}, func(err error) {
		cancel()
	})

	// Listen for termination signals.
	g.Add(run.SignalHandler(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		if *logLevel == "debug" {
			// Use %+v for github.com/pkg/errors error to print with stack.
			level.Error(logger).Log("err", fmt.Sprintf("%+v", fmt.Errorf("%s command failed %w", cmd, err)))
			os.Exit(1)
		}
		level.Error(logger).Log("err", fmt.Errorf("%s command failed %w", cmd, err))
		os.Exit(1)
	}
}

func registerLogin(_ context.Context, app *extkingpin.App) {
	cmd := app.Command("login", "Login as a tenant. Will also save tenant details locally.")
	_ = cmd.Flag("tenant", "The name of the tenant.").String()
	_ = cmd.Flag("observatorium-api-url", "The URL of the Observatorium API.").URL()
	_ = extflag.RegisterPathOrContent(cmd, "observatorium-ca", "the TLS CA against which to verify the Observatorium API. If no server CA is specified, the client will use the system certificates.")
	_ = cmd.Flag("oidc.issuer-url", "The OIDC issuer URL, see https://openid.net/specs/openid-connect-discovery-1_0.html#IssuerDiscovery.").URL()
	_ = cmd.Flag("oidc.client-secret", "The OIDC client secret, see https://tools.ietf.org/html/rfc6749#section-2.3.").String()
	_ = cmd.Flag("oidc.client-id", "The OIDC client ID, see https://tools.ietf.org/html/rfc6749#section-2.3.").String()
	_ = cmd.Flag("oidc.audience", "The audience for whom the access token is intended, see https://openid.net/specs/openid-connect-core-1_0.html#IDToken.").String()

	cmd.Run(func(ctx context.Context, logger log.Logger) error {
		return nil
	})
}

func registerCurrent(_ context.Context, app *extkingpin.App) {
	cmd := app.Command("current", "Display configuration for the currently logged in tenant.")

	cmd.Run(func(ctx context.Context, logger log.Logger) error {
		return nil
	})
}

func registerSwitch(_ context.Context, app *extkingpin.App) {
	cmd := app.Command("switch", "Switch to another locally saved tenant.")
	_ = cmd.Arg("tenant-name", "Name of tenant to switch to.").String()

	cmd.Run(func(ctx context.Context, logger log.Logger) error {
		return nil
	})
}

func registerLogout(_ context.Context, app *extkingpin.App) {
	cmd := app.Command("logout", "Logout currently logged in tenant.")

	cmd.Run(func(ctx context.Context, logger log.Logger) error {
		return nil
	})
}

func registerRead(_ context.Context, app *extkingpin.App) {
	cmd := app.Command("read", "Read series, labels & rules of a tenant.")
	_ = cmd.Flag("series", "Get series of a tenant.").Bool()
	_ = cmd.Flag("labels", "Get labels of a tenant.").Bool()
	_ = cmd.Flag("rules", "Get rules of a tenant.").Bool()

	cmd.Run(func(ctx context.Context, logger log.Logger) error {
		return nil
	})
}

func registerRules(_ context.Context, app *extkingpin.App) {
	cmd := app.Command("rules", "Read/write Prometheus Rules configuration for a tenant.")
	_ = extflag.RegisterPathOrContent(cmd, "set", "Rules configuration which will be set for a tenant.")
	_ = cmd.Flag("get", "Get configured rules in YAML form for a tenant.").Bool()

	cmd.Run(func(ctx context.Context, logger log.Logger) error {
		return nil
	})
}

func registerQuery(_ context.Context, app *extkingpin.App) {
	cmd := app.Command("query", "Query metrics for a tenant.")
	_ = cmd.Arg("query", "PromQL query for which to fetch results.").Required().String()

	cmd.Run(func(ctx context.Context, logger log.Logger) error {
		return nil
	})
}
