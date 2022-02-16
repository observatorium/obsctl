package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Read/write Prometheus Rules configuration for a tenant.",
	Long:  "Read/write Prometheus Rules configuration for a tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "rules called")
	},
}

func init() {
	rootCmd.AddCommand(rulesCmd)

	rulesCmd.Flags().String("set", "", "Path to Rules configuration file, which will be set for a tenant.")
	rulesCmd.Flags().Bool("get", false, "Get configured rules in YAML form for a tenant.")
}
