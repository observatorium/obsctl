package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "View/Add/Edit context configuration.",
	Long:  "View/Add/Edit context configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "context called")
	},
}

var contextApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Add/edit API configuration.",
	Long:  "Add/edit API configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "api called")
	},
}
var contextSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch to another context.",
	Long:  "View/Add/Edit context configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "switch called")
	},
}
var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "View current context configuration.",
	Long:  "View current context configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "current called")
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)

	contextCmd.AddCommand(contextApiCmd)
	contextCmd.AddCommand(contextSwitchCmd)
	contextCmd.AddCommand(contextCurrentCmd)
}
