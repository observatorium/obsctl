package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Display configuration for the currently logged in tenant.",
	Long:  "Display configuration for the currently logged in tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "current called")
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
