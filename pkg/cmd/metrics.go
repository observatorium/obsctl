package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/log/level"
	"github.com/observatorium/obsctl/pkg/fetcher"
	"github.com/spf13/cobra"
)

// TODO(saswatamcode): Add flags for URL query params.
func NewMetricsGetCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read series, labels & rules (JSON/YAML) of a tenant.",
		Long:  "Read series, labels & rules (JSON/YAML) of a tenant.",
	}

	// Series command.
	var seriesMatchers []string
	var seriesStart, seriesEnd string
	seriesCmd := &cobra.Command{
		Use:   "series",
		Short: "Get series of a tenant.",
		Long:  "Get series of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &fetcher.GetSeriesParams{}
			if len(seriesMatchers) > 0 {
				matcher := fetcher.SeriesMatcher(seriesMatchers)
				params.Match = &matcher
			}
			if seriesStart != "" {
				start := fetcher.StartTS(seriesStart)
				params.Start = &start
			}
			if seriesEnd != "" {
				end := fetcher.EndTS(seriesEnd)
				params.End = &end
			}

			resp, err := f.GetSeriesWithResponse(ctx, currentTenant, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return prettyPrintJSON(resp.Body)
		},
	}
	seriesCmd.Flags().StringArrayVarP(&seriesMatchers, "match", "m", nil, "Repeated series selector argument that selects the series to return.")
	seriesCmd.Flags().StringVarP(&seriesStart, "start", "s", "", "Start timestamp.")
	seriesCmd.Flags().StringVarP(&seriesEnd, "end", "e", "", "End timestamp.")

	// Labels command.
	var labelMatchers []string
	var labelStart, labelEnd string
	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "Get labels of a tenant.",
		Long:  "Get labels of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &fetcher.GetLabelsParams{}
			if len(labelMatchers) > 0 {
				matcher := fetcher.SeriesMatcher(labelMatchers)
				params.Match = &matcher
			}
			if labelStart != "" {
				start := fetcher.StartTS(labelStart)
				params.Start = &start
			}
			if labelEnd != "" {
				end := fetcher.EndTS(labelEnd)
				params.End = &end
			}

			resp, err := f.GetLabelsWithResponse(ctx, currentTenant, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return prettyPrintJSON(resp.Body)
		},
	}
	labelsCmd.Flags().StringArrayVarP(&labelMatchers, "match", "m", []string{}, "Repeated series selector argument that selects the series from which to read the label names.")
	labelsCmd.Flags().StringVarP(&labelStart, "start", "s", "", "Start timestamp.")
	labelsCmd.Flags().StringVarP(&labelEnd, "end", "e", "", "End timestamp.")

	// Labelvalues command.
	var labelValuesMatchers []string
	var labelName, labelValuesStart, labelValuesEnd string
	labelValuesCmd := &cobra.Command{
		Use:   "labelvalues",
		Short: "Get label values of a tenant.",
		Long:  "Get label values of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &fetcher.GetLabelValuesParams{}
			if len(labelValuesMatchers) > 0 {
				matcher := fetcher.SeriesMatcher(labelValuesMatchers)
				params.Match = &matcher
			}
			if labelValuesStart != "" {
				start := fetcher.StartTS(labelValuesStart)
				params.Start = &start
			}
			if labelValuesEnd != "" {
				end := fetcher.EndTS(labelValuesEnd)
				params.End = &end
			}

			resp, err := f.GetLabelValuesWithResponse(ctx, currentTenant, labelName, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return prettyPrintJSON(resp.Body)
		},
	}
	labelValuesCmd.Flags().StringVar(&labelName, "name", "", "Name of the label to fetch values for.")
	labelValuesCmd.Flags().StringArrayVarP(&labelValuesMatchers, "match", "m", []string{}, "Repeated series selector argument that selects the series from which to read the label values.")
	labelValuesCmd.Flags().StringVarP(&labelValuesStart, "start", "s", "", "Start timestamp.")
	labelValuesCmd.Flags().StringVarP(&labelValuesEnd, "end", "e", "", "End timestamp.")

	err := labelValuesCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}

	// Rules command.
	var ruleMatchers []string
	var ruleType string
	rulesCmd := &cobra.Command{
		Use:   "rules",
		Short: "Get rules of a tenant.",
		Long:  "Get rules of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &fetcher.GetRulesParams{}
			if len(ruleMatchers) > 0 {
				params.Match = &ruleMatchers
			}
			if ruleType != "" {
				if ruleType != "alert" && ruleType != "record" {
					return fmt.Errorf("not valid rule type")
				}
				params.Type = &ruleType
			}

			resp, err := f.GetRulesWithResponse(ctx, currentTenant, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return prettyPrintJSON(resp.Body)
		},
	}
	rulesCmd.Flags().StringArrayVarP(&ruleMatchers, "match", "m", []string{}, "Repeated series selector argument that selects the series from which to read the label values.")
	rulesCmd.Flags().StringVarP(&ruleType, "type", "t", "", "Rule type to filter by i.e, alert or record. No filtering done if skipped.")

	// Rules raw command.
	rulesRawCmd := &cobra.Command{
		Use:   "rules.raw",
		Short: "Get configured rules of a tenant.",
		Long:  "Get configured rules of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			resp, err := f.GetRawRulesWithResponse(ctx, currentTenant)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			fmt.Fprintln(os.Stdout, string(resp.Body))
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

func NewMetricsSetCmd(ctx context.Context) *cobra.Command {
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

			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			resp, err := f.SetRawRulesWithBodyWithResponse(ctx, currentTenant, "application/yaml", file)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			fmt.Fprintln(os.Stdout, string(resp.Body))
			return nil
		},
	}

	cmd.Flags().StringVar(&ruleFilePath, "rule.file", "", "Path to Rules configuration file, which will be set for a tenant.")
	err := cmd.MarkFlagRequired("rule.file")
	if err != nil {
		panic(err)
	}

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

func NewMetricsCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Metrics based operations for Observatorium.",
		Long:  "Metrics based operations for Observatorium.",
	}

	cmd.AddCommand(NewMetricsGetCmd(ctx))
	cmd.AddCommand(NewMetricsSetCmd(ctx))
	cmd.AddCommand(NewMetricsQueryCmd(ctx))

	return cmd
}
