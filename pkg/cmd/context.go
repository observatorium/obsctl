package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/observatorium/obsctl/pkg/config"
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
		Long:  "Add API configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.Read(logger)
			if err != nil {
				return err
			}

			return conf.AddAPI(logger, config.APIName(addName), addURL)
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
			conf, err := config.Read(logger)
			if err != nil {
				return err
			}

			return conf.RemoveAPI(logger, config.APIName(rmName))
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
			cntxt := strings.Split(args[0], "/")
			if len(cntxt) != 2 {
				return fmt.Errorf("invalid context name: use format <api>/<tenant>")
			}

			conf, err := config.Read(logger)
			if err != nil {
				return err
			}

			return conf.SetCurrent(logger, config.APIName(cntxt[0]), config.TenantName(cntxt[1]))
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "View current context configuration.",
		Long:  "View current context configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.Read(logger)
			if err != nil {
				return err
			}

			_, _, err = conf.GetCurrent()
			if err != nil {
				return err
			}

			// TODO(saswatamcode): Add flag to display more details. Eg -verbose
			fmt.Fprintf(os.Stdout, "The current context is: %s/%s\n", conf.Current.API, conf.Current.Tenant)
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "View all context configuration.",
		Long:  "View all context configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.Read(logger)
			if err != nil {
				return err
			}

			for k, v := range conf.APIs {
				fmt.Fprintf(os.Stdout, "%s\n", string(k))
				for kc := range v.Contexts {
					fmt.Fprintf(os.Stdout, "\t- %s\n", string(kc))
				}
			}

			_, _, err = conf.GetCurrent()
			if err != nil {
				return err
			}

			// TODO(saswatamcode): Add flag to display more details. Eg -verbose
			fmt.Fprintf(os.Stdout, "\nThe current context is: %s/%s\n", conf.Current.API, conf.Current.Tenant)
			return nil
		},
	}

	cmd.AddCommand(apiCmd)
	cmd.AddCommand(switchCmd)
	cmd.AddCommand(currentCmd)
	cmd.AddCommand(listCmd)

	apiCmd.AddCommand(apiAddCmd)
	apiCmd.AddCommand(apiRmCmd)

	return cmd
}
