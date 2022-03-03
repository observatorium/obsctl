package cmd

import (
	"context"

	"github.com/observatorium/obsctl/pkg/config"
	"github.com/spf13/cobra"
)

func NewLogoutCmd(ctx context.Context) *cobra.Command {
	var tenantName, apiName string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout a tenant. Will remove locally saved details.",
		Long:  "Logout a tenant. Will remove locally saved details.",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.Read(logger)
			if err != nil {
				return err
			}

			// If only one API is saved, we can assume tenant belongs to that API.
			if len(conf.APIs) == 1 {
				for k := range conf.APIs {
					return conf.RemoveTenant(logger, config.TenantName(tenantName), k)
				}
			}

			return conf.RemoveTenant(logger, config.TenantName(tenantName), config.APIName(apiName))
		},
	}

	cmd.Flags().StringVar(&tenantName, "tenant", "", "The name of the tenant to logout.")
	cmd.Flags().StringVar(&apiName, "api", "", "The name of the API the tenant is associated with. Not needed in case only one API is saved locally.")

	err := cmd.MarkFlagRequired("tenant")
	if err != nil {
		panic(err)
	}

	return cmd
}
