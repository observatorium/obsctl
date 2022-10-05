package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/observatorium/api/client"
	"github.com/observatorium/api/client/parameters"
	"github.com/observatorium/obsctl/pkg/fetcher"
	"github.com/spf13/cobra"
)

func NewLogsGetCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read series, labels & labels values (JSON/YAML) of a tenant.",
		Long:  "Read series, labels & labels values (JSON/YAML) of a tenant.",
	}

	// Series command.
	var (
		seriesMatchers         []string
		seriesStart, seriesEnd string
	)
	seriesCmd := &cobra.Command{
		Use:          "series",
		Short:        "Get series of a tenant.",
		Long:         "Get series of a tenant.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &client.GetSeriesParams{}
			if len(seriesMatchers) > 0 {
				params.Match = seriesMatchers
			}
			if seriesStart != "" {
				params.Start = (*parameters.StartTS)(&seriesStart)
			}
			if seriesEnd != "" {
				params.End = (*parameters.EndTS)(&seriesEnd)
			}

			resp, err := f.GetSeriesWithResponse(ctx, currentTenant, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	seriesCmd.Flags().StringArrayVarP(&seriesMatchers, "match", "m", nil, "Repeated series selector argument that selects the series to return.")
	seriesCmd.Flags().StringVarP(&seriesStart, "start", "s", "", "Start timestamp.")
	seriesCmd.Flags().StringVarP(&seriesEnd, "end", "e", "", "End timestamp.")
	err := seriesCmd.MarkFlagRequired("match")
	if err != nil {
		panic(err)
	}

	// Labels command.
	var (
		labelStart, labelEnd string
	)
	labelsCmd := &cobra.Command{
		Use:   "labels",
		Short: "Get labels of a tenant.",
		Long:  "Get labels of a tenant.",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &client.GetLogLabelsParams{}
			if labelStart != "" {
				params.Start = (*parameters.StartTS)(&labelStart)
			}
			if labelEnd != "" {
				params.End = (*parameters.EndTS)(&labelEnd)
			}

			resp, err := f.GetLogLabelsWithResponse(ctx, currentTenant, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}

	labelsCmd.Flags().StringVarP(&labelStart, "start", "s", "", "Start timestamp.")
	labelsCmd.Flags().StringVarP(&labelEnd, "end", "e", "", "End timestamp.")

	// Labelvalues command.
	var (
		labelName, labelValuesStart, labelValuesEnd string
	)
	labelValuesCmd := &cobra.Command{
		Use:          "labelvalues",
		Short:        "Get label values of a tenant.",
		Long:         "Get label values of a tenant.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			params := &client.GetLogLabelValuesParams{}

			if labelValuesStart != "" {
				params.Start = (*parameters.StartTS)(&labelValuesStart)
			}
			if labelValuesEnd != "" {
				params.End = (*parameters.EndTS)(&labelValuesEnd)
			}

			resp, err := f.GetLogLabelValuesWithResponse(ctx, currentTenant, labelName, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	labelValuesCmd.Flags().StringVar(&labelName, "name", "", "Name of the label to fetch values for.")
	labelValuesCmd.Flags().StringVarP(&labelValuesStart, "start", "s", "", "Start timestamp.")
	labelValuesCmd.Flags().StringVarP(&labelValuesEnd, "end", "e", "", "End timestamp.")

	err = labelValuesCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}

	// Alerts Command.
	alertsCmd := &cobra.Command{
		Use:          "alerts",
		Short:        "Get alerts of a tenant.",
		Long:         "Get alerts of a tenant.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			resp, err := f.GetLogsPromAlertsWithResponse(ctx, currentTenant)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}

	// Rules command.
	rulesCmd := &cobra.Command{
		Use:          "rules",
		Short:        "Get rules of a tenant.",
		Long:         "Get rules of a tenant.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			resp, err := f.GetLogsPromRulesWithResponse(ctx, currentTenant)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}

	// Rules.raw command.
	var (
		rulesNamespace, rulesGroup string
	)
	rulesRawCmd := &cobra.Command{
		Use:          "rules.raw",
		Short:        "Get configured rules of a tenant.",
		Long:         "Get configured rules of a tenant.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			if rulesNamespace == "" {
				resp, err := f.GetAllLogsRulesWithResponse(ctx, currentTenant)
				if err != nil {
					return fmt.Errorf("getting response: %w", err)
				}

				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
			}

			if rulesGroup != "" {
				resp, err := f.GetLogsRulesGroupWithResponse(
					ctx, currentTenant, parameters.LogRulesNamespace(rulesNamespace), parameters.LogRulesGroup(rulesGroup),
				)
				if err != nil {
					return fmt.Errorf("getting response: %w", err)
				}

				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
			}

			resp, err := f.GetLogsRulesWithResponse(ctx, currentTenant, parameters.LogRulesNamespace(rulesNamespace))
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	rulesRawCmd.Flags().StringVarP(&rulesNamespace, "namespace", "n", "", "Rules Namespace")
	rulesRawCmd.Flags().StringVarP(&rulesGroup, "group", "g", "", "Rules Group in a namespace")

	cmd.AddCommand(seriesCmd)
	cmd.AddCommand(labelsCmd)
	cmd.AddCommand(labelValuesCmd)
	cmd.AddCommand(alertsCmd)
	cmd.AddCommand(rulesCmd)
	cmd.AddCommand(rulesRawCmd)

	return cmd
}

func NewLogsSetCmd(ctx context.Context) *cobra.Command {
	var (
		rulesNamespace, ruleFilePath string
	)
	cmd := &cobra.Command{
		Use:          "set",
		Short:        "Write Loki Rules configuration for a tenant.",
		Long:         "Write Loki Rules configuration for a tenant.",
		SilenceUsage: true,
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

			resp, err := f.SetLogsRulesWithBodyWithResponse(ctx, currentTenant, parameters.LogRulesNamespace(rulesNamespace), "application/yaml", file)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			if resp.StatusCode()/100 != 2 {
				if len(resp.Body) != 0 {
					fmt.Fprintln(cmd.OutOrStdout(), string(resp.Body))
					return fmt.Errorf("request failed with status code %d", resp.StatusCode())
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(resp.Body))
			return nil
		},
	}

	cmd.Flags().StringVarP(&rulesNamespace, "namespace", "n", "", "Rules Namespace")
	cmd.Flags().StringVar(&ruleFilePath, "rule.file", "", "Path to Rules configuration file, which will be set for a tenant.")

	err := cmd.MarkFlagRequired("rule.file")
	if err != nil {
		panic(err)
	}

	err = cmd.MarkFlagRequired("namespace")
	if err != nil {
		panic(err)
	}

	return cmd
}

func NewLogsQueryCmd(ctx context.Context) *cobra.Command {
	var (
		isRange                                     bool
		time, start, end, direction, step, interval string
		limit                                       float32
	)
	cmd := &cobra.Command{
		Use:          "query",
		Short:        "Query logs for a tenant.",
		Long:         "Query logs for a tenant. Can get results for both instant and range queries. Pass a single valid LogQl query to fetch results for.",
		Example:      `obsctl logs query "prometheus_http_request_total"`,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "" {
				return fmt.Errorf("no query provided")
			}

			f, currentTenant, err := fetcher.NewCustomFetcher(ctx, logger)
			if err != nil {
				return fmt.Errorf("custom fetcher: %w", err)
			}

			query := parameters.LogqlQuery(args[0])

			if isRange {
				params := &client.GetLogRangeQueryParams{Query: &query}
				if limit != 0 {
					params.Limit = (*parameters.Limit)(&limit)
				}

				if start == "" || end == "" {
					return fmt.Errorf("start/end timestamp not provided for range query")
				}

				params.Start = (*parameters.StartTS)(&start)
				params.End = (*parameters.EndTS)(&end)

				if step != "" {
					params.Step = &step
				}

				if interval != "" {
					params.Interval = &interval
				}

				if direction != "" {
					params.Direction = &direction
				}

				resp, err := f.GetLogRangeQueryWithResponse(ctx, currentTenant, params)
				if err != nil {
					return fmt.Errorf("getting response: %w", err)
				}

				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
			} else {
				params := &client.GetLogInstantQueryParams{Query: &query}
				if time != "" {
					params.Time = &time
				}

				if limit != 0 {
					params.Limit = (*parameters.Limit)(&limit)
				}

				if direction != "" {
					params.Direction = &direction
				}

				resp, err := f.GetLogInstantQueryWithResponse(ctx, currentTenant, params)
				if err != nil {
					return fmt.Errorf("getting response: %w", err)
				}

				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
			}
		},
	}

	// Flags for instant query.
	cmd.Flags().StringVar(&time, "time", "", "Evaluation timestamp. Only used if --range is false.")

	// Flags for range query.
	cmd.Flags().BoolVar(&isRange, "range", false, "If true, query will be evaluated as a range query. See https://prometheus.io/docs/prometheus/latest/querying/api/#range-queries.")
	cmd.Flags().StringVarP(&start, "start", "s", "", "Start timestamp. Must be provided if --range is true.")
	cmd.Flags().StringVarP(&end, "end", "e", "", "End timestamp. Must be provided if --range is true.")
	cmd.Flags().StringVar(&step, "step", "", "Query resolution step width. Only used if --range is provided.")
	cmd.Flags().StringVar(&interval, "interval", "", "return entries at (or greater than) the specified interval,Only used if --range is provided.")

	// // Common flags.
	cmd.Flags().Float32Var(&limit, "limit", 100, "The max number of entries to return. Only used if --range is false.")
	cmd.Flags().StringVar(&direction, "direction", "", "Determines the sort order of logs.. Only used if --range is false.")

	return cmd
}

func NewLogsCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "logs based operations for Observatorium.",
		Long:  "logs based operations for Observatorium.",
	}

	cmd.AddCommand(NewLogsGetCmd(ctx))
	cmd.AddCommand(NewLogsSetCmd(ctx))
	cmd.AddCommand(NewLogsQueryCmd(ctx))

	return cmd
}
