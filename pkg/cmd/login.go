package cmd

import (
	"context"
	"io/ioutil"

	"github.com/observatorium/obsctl/pkg/auth"
	"github.com/spf13/cobra"
)

func NewLoginCmd(ctx context.Context) *cobra.Command {
	var tenantCfg auth.TenantConfig
	var caFilePath string
	var disableOIDCCheck bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login as a tenant. Will also save tenant details locally.",
		Long:  "Login as a tenant. Will also save tenant details locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if caFilePath != "" {
				body, err := ioutil.ReadFile(caFilePath)
				if err != nil {
					return err
				}
				tenantCfg.CAFile = body
			}
			return auth.Login(ctx, tenantCfg, disableOIDCCheck, logger)
		},
	}

	cmd.Flags().StringVar(&tenantCfg.Tenant, "tenant", "", "The name of the tenant.")
	cmd.Flags().StringVar(&tenantCfg.APIname, "api", "", "The name of the Observatorium API.")

	cmd.Flags().StringVar(&caFilePath, "ca", "", "Path to the TLS CA against which to verify the Observatorium API. If no server CA is specified, the client will use the system certificates.")
	cmd.Flags().StringVar(&tenantCfg.OIDC.IssuerURL, "oidc.issuer-url", "", "The OIDC issuer URL, see https://openid.net/specs/openid-connect-discovery-1_0.html#IssuerDiscovery.")
	cmd.Flags().StringVar(&tenantCfg.OIDC.ClientSecret, "oidc.client-secret", "", "The OIDC client secret, see https://tools.ietf.org/html/rfc6749#section-2.3.")
	cmd.Flags().StringVar(&tenantCfg.OIDC.ClientID, "oidc.client-id", "", "The OIDC client ID, see https://tools.ietf.org/html/rfc6749#section-2.3.")
	cmd.Flags().StringVar(&tenantCfg.OIDC.Audience, "oidc.audience", "", "The audience for whom the access token is intended, see https://openid.net/specs/openid-connect-core-1_0.html#IDToken.")

	cmd.Flags().BoolVar(&disableOIDCCheck, "disable.oidc-check", false, "If set to true, OIDC flags will not be checked while saving tenant details locally.")

	err := cmd.MarkFlagRequired("api")
	if err != nil {
		panic(err)
	}

	err = cmd.MarkFlagRequired("tenant")
	if err != nil {
		panic(err)
	}

	return cmd
}
