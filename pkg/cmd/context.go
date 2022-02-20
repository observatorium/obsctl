package cmd

import (
	"context"

	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

func NewContextCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "View/Add/Edit context configuration.",
		Long:  "View/Add/Edit context configuration.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "context called")
		},
	}

	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Add/edit API configuration.",
		Long:  "Add/edit API configuration.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "api called")
		},
	}
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch to another context.",
		Long:  "View/Add/Edit context configuration.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "switch called")
		},
	}
	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "View current context configuration.",
		Long:  "View current context configuration.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "current called")
		},
	}

	cmd.AddCommand(apiCmd)
	cmd.AddCommand(switchCmd)
	cmd.AddCommand(currentCmd)

	return cmd
}
