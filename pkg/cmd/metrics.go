package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Metrics based operations for Observatorium.",
	Long:  "Metrics based operations for Observatorium.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "metrics called")
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
