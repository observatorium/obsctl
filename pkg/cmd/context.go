package cmd

import (
	"context"
	"fmt"

	"github.com/go-kit/log/level"
	"github.com/observatorium/obsctl/pkg/auth"
	"github.com/spf13/cobra"
)

func NewContextCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage context configuration.",
		Long:  "Manage context configuration.",
	}

	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Add/edit API configuration.",
		Long:  "Add/edit API configuration.",
	}

	var addURL, addName string
	apiAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Add API configuration.",
		Long:  "Add API configuration. If there is a previously saved config with the same name, it will be updated.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.AddAPI(addURL, addName, logger)
		},
	}

	apiAddCmd.Flags().StringVar(&addURL, "url", "", "The URL for the Observatorium API.")
	apiAddCmd.Flags().StringVar(&addName, "name", "", "Provide an optional name to easily refer to the Observatorium Instance.")
	err := apiAddCmd.MarkFlagRequired("url")
	if err != nil {
		panic(err)
	}

	var rmName string
	apiRmCmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove API configuration.",
		Long:  "Remove API configuration. If set to current, current will be set to nil.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.RemoveAPI(rmName, logger)
		},
	}

	apiRmCmd.Flags().StringVar(&rmName, "name", "", "The name of the Observatorium API instance to remove.")
	err = apiRmCmd.MarkFlagRequired("name")
	if err != nil {
		panic(err)
	}

	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "Switch to another context.",
		Long:  "View/Add/Edit context configuration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.SwitchContext(args[0], logger)
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "View current context configuration.",
		Long:  "View current context configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			tenantCfg, apiCfg, err := auth.GetCurrentContext()
			if err != nil {
				return err
			}

			// TODO: Add flag to display more details. Eg -verbose
			level.Info(logger).Log("msg", fmt.Sprintf("The current context is %s/%s", apiCfg.Name, tenantCfg.Tenant))
			return nil
		},
	}

	cmd.AddCommand(apiCmd)
	cmd.AddCommand(switchCmd)
	cmd.AddCommand(currentCmd)

	apiCmd.AddCommand(apiAddCmd)
	apiCmd.AddCommand(apiRmCmd)

	return cmd
}
