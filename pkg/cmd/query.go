package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:     "query",
	Short:   "Query metrics for a tenant.",
	Long:    "Query metrics for a tenant. Pass a single valid PromQL query to fetch results for.",
	Example: `obsctl query "prometheus_http_request_total"`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "query called")
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
}
