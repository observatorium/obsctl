package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:     "switch",
	Short:   "Switch to another locally saved tenant.",
	Long:    "Switch to another locally saved tenant.",
	Example: "switch <name of tenant>",
	Args:    cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "switch called")
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
