package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-kit/log/level"
	"github.com/observatorium/api/client"
	"github.com/observatorium/api/client/parameters"
	"github.com/observatorium/obsctl/pkg/fetcher"
	"github.com/observatorium/obsctl/pkg/proxy"
	"github.com/spf13/cobra"
)

func NewMetricsGetCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read series, labels & rules (JSON/YAML) of a tenant.",
		Long:  "Read series, labels & rules (JSON/YAML) of a tenant.",
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
		labelMatchers        []string
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

			params := &client.GetLabelsParams{}
			if len(labelMatchers) > 0 {
				params.Match = (*parameters.OptionalSeriesMatcher)(&labelMatchers)
			}
			if labelStart != "" {
				params.Start = (*parameters.StartTS)(&labelStart)
			}
			if labelEnd != "" {
				params.End = (*parameters.EndTS)(&labelEnd)
			}

			resp, err := f.GetLabelsWithResponse(ctx, currentTenant, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	labelsCmd.Flags().StringArrayVarP(&labelMatchers, "match", "m", []string{}, "Repeated series selector argument that selects the series from which to read the label names.")
	labelsCmd.Flags().StringVarP(&labelStart, "start", "s", "", "Start timestamp.")
	labelsCmd.Flags().StringVarP(&labelEnd, "end", "e", "", "End timestamp.")

	// Labelvalues command.
	var (
		labelValuesMatchers                         []string
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

			params := &client.GetLabelValuesParams{}
			if len(labelValuesMatchers) > 0 {
				params.Match = (*parameters.OptionalSeriesMatcher)(&labelValuesMatchers)
			}
			if labelValuesStart != "" {
				params.Start = (*parameters.StartTS)(&labelValuesStart)
			}
			if labelValuesEnd != "" {
				params.End = (*parameters.EndTS)(&labelValuesEnd)
			}

			resp, err := f.GetLabelValuesWithResponse(ctx, currentTenant, labelName, params)
			if err != nil {
				return fmt.Errorf("getting response: %w", err)
			}

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	labelValuesCmd.Flags().StringVar(&labelName, "name", "", "Name of the label to fetch values for.")
	labelValuesCmd.Flags().StringArrayVarP(&labelValuesMatchers, "match", "m", []string{}, "Repeated series selector argument that selects the series from which to read the label values.")
	labelValuesCmd.Flags().StringVarP(&labelValuesStart, "start", "s", "", "Start timestamp.")
	labelValuesCmd.Flags().StringVarP(&labelValuesEnd, "end", "e", "", "End timestamp.")

	err = labelValuesCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}

	// Rules command.
	var (
		ruleMatchers []string
		ruleType     string
	)
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

			params := &client.GetRulesParams{}
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

			return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
		},
	}
	rulesCmd.Flags().StringArrayVarP(&ruleMatchers, "match", "m", []string{}, "Repeated series selector argument that selects the series from which to read the label values.")
	rulesCmd.Flags().StringVarP(&ruleType, "type", "t", "", "Rule type to filter by i.e, alert or record. No filtering done if skipped.")

	// Rules raw command.
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

			resp, err := f.GetRawRulesWithResponse(ctx, currentTenant)
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
		Use:          "set",
		Short:        "Write Prometheus Rules configuration for a tenant.",
		Long:         "Write Prometheus Rules configuration for a tenant.",
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

			resp, err := f.SetRawRulesWithBodyWithResponse(ctx, currentTenant, "application/yaml", file)
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

	cmd.Flags().StringVar(&ruleFilePath, "rule.file", "", "Path to Rules configuration file, which will be set for a tenant.")
	err := cmd.MarkFlagRequired("rule.file")
	if err != nil {
		panic(err)
	}

	return cmd
}

func NewMetricsQueryCmd(ctx context.Context) *cobra.Command {
	var (
		isRange                                    bool
		evalTime, timeout, start, end, step, graph string
	)
	cmd := &cobra.Command{
		Use:          "query",
		Short:        "Query metrics for a tenant.",
		Long:         "Query metrics for a tenant. Can get results for both instant and range queries. Pass a single valid PromQL query to fetch results for.",
		Example:      `obsctl metrics query "prometheus_http_request_total"`,
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

			query := parameters.PromqlQuery(args[0])

			if isRange {
				params := &client.GetRangeQueryParams{Query: &query}
				if timeout != "" {
					params.Timeout = (*parameters.QueryTimeout)(&timeout)
				}

				if start == "" || end == "" {
					return fmt.Errorf("start/end timestamp not provided for range query")
				}

				params.Start = (*parameters.StartTS)(&start)
				params.End = (*parameters.EndTS)(&end)

				if step != "" {
					params.Step = &step
				}

				resp, err := f.GetRangeQueryWithResponse(ctx, currentTenant, params)
				if err != nil {
					return fmt.Errorf("getting response: %w", err)
				}

				if graph != "" {
					wd, err := os.Getwd()
					if err != nil {
						return fmt.Errorf("could not get working dir: %w", err)
					}

					return handleGraph(resp.Body, graph, string(query), path.Join(wd, "graph"+time.Now().String()+".png"), cmd.OutOrStdout())
				}

				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
			} else {
				params := &client.GetInstantQueryParams{Query: &query}
				if evalTime != "" {
					params.Time = &evalTime
				}
				if timeout != "" {
					params.Timeout = (*parameters.QueryTimeout)(&timeout)
				}

				resp, err := f.GetInstantQueryWithResponse(ctx, currentTenant, params)
				if err != nil {
					return fmt.Errorf("getting response: %w", err)
				}

				return handleResponse(resp.Body, resp.HTTPResponse.Header.Get("content-type"), resp.StatusCode(), cmd)
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
	cmd.Flags().StringVar(&graph, "graph", "", "If specified, range query result will output an (ascii|png) graph.")

	// Common flags.
	cmd.Flags().StringVar(&timeout, "timeout", "", "Evaluation timeout. Optional.")

	return cmd
}

func NewMetricsUICmd(ctx context.Context) *cobra.Command {
	var listen string
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Starts a proxy server and opens a Thanos Query UI for making requests to Observatorium API as a tenant.",
		Long: `Starts a proxy server and opens a Thanos Query UI for making requests to Observatorium API as a tenant. 
		Note that all request URLs will have /api/metrics/v1/ prefixed in their paths.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Run a server as we would in main.
			s, err := proxy.NewProxyServer(ctx, logger, "metrics", listen)
			if err != nil {
				return err
			}

			go func() {
				level.Info(logger).Log("msg", "starting ui proxy server", "addr", listen)

				if err = s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					level.Error(logger).Log("msg", "failed to start proxy server", "error", err)
				}
			}()

			// Open Querier UI in browser.
			if err := openInBrowser("http://localhost" + listen); err != nil {
				return err
			}

			<-ctx.Done()
			return s.Shutdown(context.Background())
		},
	}

	cmd.Flags().StringVar(&listen, "listen", ":8080", "Address for proxy server to listen on.")

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
	cmd.AddCommand(NewMetricsUICmd(ctx))

	return cmd
}
