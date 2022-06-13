package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-kit/log/level"
	"github.com/observatorium/obsctl/pkg/config"
	"github.com/spf13/cobra"
)

func NewTraceServicesCmd(ctx context.Context) *cobra.Command {
	var outputFormat string
	cmd := &cobra.Command{
		Use:   "services",
		Short: "The names of services with trace information",
		Long:  "The names of services with trace information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Read(logger)
			if err != nil {
				return fmt.Errorf("getting reading config: %w", err)
			}

			client, err := cfg.Client(ctx, logger)
			if err != nil {
				return fmt.Errorf("getting current client: %w", err)
			}

			level.Debug(logger).Log(
				"msg", "Using configuration",
				"URL", cfg.APIs[cfg.Current.API].URL,
				"tenant", cfg.Current.Tenant)

			url := fmt.Sprintf("%s%s", cfg.APIs[cfg.Current.API].URL, "api/traces/v1/rhobs/api/services")
			resp, err := client.Get(url)
			if err != nil {
				return fmt.Errorf("getting: %w", err)
			}
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("getting: %w", err)
			}
			if resp.StatusCode >= 300 {
				fmt.Fprintf(os.Stdout, "%d: %s\n%s", resp.StatusCode, resp.Status, string(bodyBytes))
				return fmt.Errorf("%d: %s", resp.StatusCode, resp.Status)
			}

			if outputFormat == "table" {
				return printTable(bodyBytes)
			} else if outputFormat == "json" {
				fmt.Printf("%s\n", string(bodyBytes))
			} else {
				return fmt.Errorf("unknown format %s", outputFormat)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: json|table")

	return cmd
}

func NewTracesCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traces",
		Short: "Trace-based operations for Observatorium.",
		Long:  "Trace-based operations for Observatorium.",
	}

	cmd.AddCommand(NewTraceServicesCmd(ctx))

	return cmd
}

func printTable(js []byte) error {
	var result map[string]interface{}
	err := json.Unmarshal(js, &result)
	if err != nil {
		return err
	}
	data, ok := result["data"]
	if !ok {
		return fmt.Errorf("no JSON data in %s", string(js))
	}
	services, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("expected JSON list in %s", string(js))
	}
	for _, svc := range services {
		fmt.Printf("%s\n", svc)
	}
	return nil
}
