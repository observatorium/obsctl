package cmd

import (
	"context"
	"fmt"

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

	cmd.AddCommand(seriesCmd)
	cmd.AddCommand(labelsCmd)
	cmd.AddCommand(labelValuesCmd)

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
	cmd.AddCommand(NewLogsQueryCmd(ctx))

	return cmd
}
