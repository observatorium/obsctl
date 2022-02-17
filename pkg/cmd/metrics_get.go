package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var metricsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Read series, labels & rules (JSON/YAML) of a tenant.",
	Long:  "Read series, labels & rules (JSON/YAML) of a tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "get called")
	},
}

var metricsGetSeriesCmd = &cobra.Command{
	Use:   "series",
	Short: "Get series of a tenant.",
	Long:  "Get series of a tenant..",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "series called")
	},
}

var metricsGetLabelsCmd = &cobra.Command{
	Use:   "labels",
	Short: "Get labels of a tenant.",
	Long:  "Get labels of a tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "labels called")
	},
}

var metricsGetRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Get rules of a tenant.",
	Long:  "Get rules of a tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "rules called")
	},
}

var metricsGetRulesRawCmd = &cobra.Command{
	Use:   "rules.raw",
	Short: "Get configured rules of a tenant.",
	Long:  "Get configured rules of a tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "rules.raw called")
	},
}

func init() {
	metricsCmd.AddCommand(metricsGetCmd)

	metricsGetCmd.AddCommand(metricsGetSeriesCmd)
	metricsGetCmd.AddCommand(metricsGetLabelsCmd)
	metricsGetCmd.AddCommand(metricsGetRulesCmd)
	metricsGetCmd.AddCommand(metricsGetRulesRawCmd)
}
