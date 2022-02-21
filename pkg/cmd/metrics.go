package cmd

import (
	"context"

	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

func NewMetricsGetCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read series, labels & rules (JSON/YAML) of a tenant.",
		Long:  "Read series, labels & rules (JSON/YAML) of a tenant.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "get called")
		},
	}

	seriesCmd := &cobra.Command{
		Use:   "series",
		Short: "Get series of a tenant.",
		Long:  "Get series of a tenant..",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "series called")
		},
	}

	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "Get labels of a tenant.",
		Long:  "Get labels of a tenant.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "labels called")
		},
	}

	labelValuesCmd := &cobra.Command{
		Use:   "labelvalues",
		Short: "Get label values of a tenant.",
		Long:  "Get label values of a tenant.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "label values called")
		},
	}

	rulesCmd := &cobra.Command{
		Use:   "rules",
		Short: "Get rules of a tenant.",
		Long:  "Get rules of a tenant.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "rules called")
		},
	}

	rulesRawCmd := &cobra.Command{
		Use:   "rules.raw",
		Short: "Get configured rules of a tenant.",
		Long:  "Get configured rules of a tenant.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "rules.raw called")
		},
	}

	cmd.AddCommand(seriesCmd)
	cmd.AddCommand(labelsCmd)
	cmd.AddCommand(labelValuesCmd)
	cmd.AddCommand(rulesCmd)
	cmd.AddCommand(rulesRawCmd)

	return cmd
}

func NewMetricsSetCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Write Prometheus Rules configuration for a tenant.",
		Long:  "Write Prometheus Rules configuration for a tenant.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "set called")
		},
	}

	cmd.Flags().String("rule.file", "", "Path to Rules configuration file, which will be set for a tenant.")

	return cmd
}

func NewMetricsQueryCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Query metrics for a tenant.",
		Long:    "Query metrics for a tenant. Pass a single valid PromQL query to fetch results for.",
		Example: `obsctl query "prometheus_http_request_total"`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "query called")
		},
	}

	return cmd
}

func NewMetricsCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Metrics based operations for Observatorium.",
		Long:  "Metrics based operations for Observatorium.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "metrics called")
		},
	}

	cmd.AddCommand(NewMetricsGetCmd(ctx))
	cmd.AddCommand(NewMetricsSetCmd(ctx))
	cmd.AddCommand(NewMetricsQueryCmd(ctx))

	return cmd
}
