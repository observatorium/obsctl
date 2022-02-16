package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read series, labels & rules of a tenant.",
	Long:  "Read series, labels & rules of a tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "read called")
	},
}

func init() {
	rootCmd.AddCommand(readCmd)

	readCmd.Flags().Bool("series", false, "Get series of a tenant.")
	readCmd.Flags().Bool("labels", false, "Get series of a tenant.")
	readCmd.Flags().Bool("rules", false, "Get series of a tenant.")
}
