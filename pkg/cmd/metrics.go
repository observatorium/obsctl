package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-kit/log/level"
	"github.com/observatorium/obsctl/pkg/config"
	"github.com/spf13/cobra"
)

// TODO(saswatamcode): Add flags for URL query params.
func NewMetricsGetCmd(ctx context.Context, path ...string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read series, labels & rules (JSON/YAML) of a tenant.",
		Long:  "Read series, labels & rules (JSON/YAML) of a tenant.",
	}

	seriesCmd := &cobra.Command{
		Use:   "series",
		Short: "Get series of a tenant.",
		Long:  "Get series of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := config.DoMetricsGetReq(ctx, logger, "/api/v1/series", path...)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "this should work")

			return prettyPrintJSON(b, cmd.OutOrStdout())
		},
	}

	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "Get labels of a tenant.",
		Long:  "Get labels of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := config.DoMetricsGetReq(ctx, logger, "/api/v1/labels", path...)
			if err != nil {
				return err
			}

			return prettyPrintJSON(b, cmd.OutOrStdout())
		},
	}

	var labelName string
	labelValuesCmd := &cobra.Command{
		Use:   "labelvalues",
		Short: "Get label values of a tenant.",
		Long:  "Get label values of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := config.DoMetricsGetReq(ctx, logger, "/api/v1/label/"+labelName+"/values", path...)
			if err != nil {
				return err
			}

			return prettyPrintJSON(b, cmd.OutOrStdout())
		},
	}
	labelValuesCmd.Flags().StringVar(&labelName, "name", "", "Name of the label to fetch values for.")
	err := labelValuesCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}

	rulesCmd := &cobra.Command{
		Use:   "rules",
		Short: "Get rules of a tenant.",
		Long:  "Get rules of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := config.DoMetricsGetReq(ctx, logger, "/api/v1/rules", path...)
			if err != nil {
				return err
			}

			return prettyPrintJSON(b, cmd.OutOrStdout())
		},
	}

	rulesRawCmd := &cobra.Command{
		Use:   "rules.raw",
		Short: "Get configured rules of a tenant.",
		Long:  "Get configured rules of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := config.DoMetricsGetReq(ctx, logger, "/api/v1/rules/raw", path...)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}

	cmd.AddCommand(seriesCmd)
	cmd.AddCommand(labelsCmd)
	cmd.AddCommand(labelValuesCmd)
	cmd.AddCommand(rulesCmd)
	cmd.AddCommand(rulesRawCmd)

	return cmd
}

func NewMetricsSetCmd(ctx context.Context, path ...string) *cobra.Command {
	var ruleFilePath string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Write Prometheus Rules configuration for a tenant.",
		Long:  "Write Prometheus Rules configuration for a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.Open(ruleFilePath)
			if err != nil {
				return fmt.Errorf("opening rule file: %w", err)
			}
			defer file.Close()

			data, err := ioutil.ReadAll(file)
			if err != nil {
				return fmt.Errorf("reading rule file: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(data))

			b, err := config.DoMetricsPutReqWithYAML(ctx, logger, "/api/v1/rules/raw", data, path...)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}

	cmd.Flags().StringVar(&ruleFilePath, "rule.file", "", "Path to Rules configuration file, which will be set for a tenant.")

	return cmd
}

func NewMetricsQueryCmd(ctx context.Context, path ...string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Query metrics for a tenant.",
		Long:    "Query metrics for a tenant. Pass a single valid PromQL query to fetch results for.",
		Example: `obsctl query "prometheus_http_request_total"`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "query not implemented yet")
		},
	}

	return cmd
}

func NewMetricsCmd(ctx context.Context, path ...string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Metrics based operations for Observatorium.",
		Long:  "Metrics based operations for Observatorium.",
	}

	cmd.AddCommand(NewMetricsGetCmd(ctx, path...))
	cmd.AddCommand(NewMetricsSetCmd(ctx, path...))
	cmd.AddCommand(NewMetricsQueryCmd(ctx, path...))

	return cmd
}
