package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var metricsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Write Prometheus Rules configuration for a tenant.",
	Long:  "Write Prometheus Rules configuration for a tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "set called")
	},
}

func init() {
	metricsCmd.AddCommand(metricsSetCmd)

	metricsSetCmd.Flags().String("rule.file", "", "Path to Rules configuration file, which will be set for a tenant.")
}
