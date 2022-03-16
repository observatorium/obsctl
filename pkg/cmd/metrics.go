package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/observatorium/obsctl/pkg/fetcher"
	"github.com/spf13/cobra"
)

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
				params.Match = (*fetcher.SeriesMatcher)(&seriesMatchers)
			}
			if seriesStart != "" {
				params.Start = (*fetcher.StartTS)(&seriesStart)
			}
			if seriesEnd != "" {
				params.End = (*fetcher.EndTS)(&seriesEnd)
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
				params.Match = (*fetcher.SeriesMatcher)(&labelMatchers)
			}
			if labelStart != "" {
				params.Start = (*fetcher.StartTS)(&labelStart)
			}
			if labelEnd != "" {
				params.End = (*fetcher.EndTS)(&labelEnd)
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
				params.Match = (*fetcher.SeriesMatcher)(&labelValuesMatchers)
			}
			if labelValuesStart != "" {
				params.Start = (*fetcher.StartTS)(&labelValuesStart)
			}
			if labelValuesEnd != "" {
				params.End = (*fetcher.EndTS)(&labelValuesEnd)
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

func NewMetricsQueryCmd(ctx context.Context) *cobra.Command {
	var isRange bool
	var evalTime, timeout, start, end, step string
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Query metrics for a tenant.",
		Long:    "Query metrics for a tenant. Can get results for both instant and range queries. Pass a single valid PromQL query to fetch results for.",
		Example: `obsctl query "prometheus_http_request_total"`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "" {
				return fmt.Errorf("no query provided")
			}

			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			query := fetcher.Query(args[0])

			if isRange {
				params := &fetcher.GetRangeQueryParams{Query: &query}
				if timeout != "" {
					params.Timeout = (*fetcher.QueryTimeout)(&timeout)
				}

				if start == "" || end == "" {
					return fmt.Errorf("start/end timestamp not provided for range query")
				}

				params.Start = (*fetcher.StartTS)(&start)
				params.End = (*fetcher.EndTS)(&end)

				if step != "" {
					params.Step = &step
				}

				resp, err := f.GetRangeQueryWithResponse(ctx, currentTenant, params)
				if err != nil {
					return fmt.Errorf("getting response: %w", err)
				}

				return prettyPrintJSON(resp.Body)
			} else {
				params := &fetcher.GetInstantQueryParams{Query: &query}
				if evalTime != "" {
					params.Time = &evalTime
				}
				if timeout != "" {
					params.Timeout = (*fetcher.QueryTimeout)(&timeout)
				}

				resp, err := f.GetInstantQueryWithResponse(ctx, currentTenant, params)
				if err != nil {
					return fmt.Errorf("getting response: %w", err)
				}

				return prettyPrintJSON(resp.Body)
			}
		},
	}

	// Flags for instant query.
	cmd.Flags().StringVar(&evalTime, "time", "", "Evaluation timestamp. Only used if --range is false.")

	// Flags for range query.
	cmd.Flags().BoolVar(&isRange, "range", false, "If true, query will be evaluated as a range query. See https://prometheus.io/docs/prometheus/latest/querying/api/#range-queries.")
	cmd.Flags().StringVarP(&start, "start", "s", "", "Start timestamp. Must be provided if --range is true.")
	cmd.Flags().StringVarP(&end, "end", "e", "", "End timestamp. Must be provided if --range is true.")
	cmd.Flags().StringVar(&step, "step", "", "Query resolution step width. Only used if --range is provided.")

	// Common flags.
	cmd.Flags().StringVar(&timeout, "timeout", "", "Evaluation timeout. Optional.")

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
