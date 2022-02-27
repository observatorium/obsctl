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
			conf, err := config.Read()
			if err != nil {
				return err
			}

			return conf.RemoveTenant(config.TenantName(tenantName), config.APIName(apiName))
		},
	}

	cmd.Flags().StringVar(&tenantName, "tenant", "", "The name of the tenant to logout.")
	cmd.Flags().StringVar(&apiName, "api", "", "The name of the API the tenant is associated with.")

	err := cmd.MarkFlagRequired("tenant")
	if err != nil {
		panic(err)
	}

	err = cmd.MarkFlagRequired("api")
	if err != nil {
		panic(err)
	}

	return cmd
}
