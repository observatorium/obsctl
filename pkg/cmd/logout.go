package cmd

import (
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout currently logged in tenant.",
	Long:  "Logout currently logged in tenant.",
	Run: func(cmd *cobra.Command, args []string) {
		level.Info(logger).Log("msg", "logout called")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
