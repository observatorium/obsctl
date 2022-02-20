package cmd

import (
	"context"

	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

func NewLoginCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login as a tenant. Will also save tenant details locally.",
		Long:  "Login as a tenant. Will also save tenant details locally.",
		Run: func(cmd *cobra.Command, args []string) {
			level.Info(logger).Log("msg", "login called")
		},
	}

	cmd.Flags().String("tenant", "", "The name of the tenant.")
	cmd.Flags().String("api", "", "The URL or name of the Observatorium API.")
	cmd.Flags().String("ca", "", "Path to the TLS CA against which to verify the Observatorium API. If no server CA is specified, the client will use the system certificates.")
	cmd.Flags().String("oidc.issuer-url", "", "The OIDC issuer URL, see https://openid.net/specs/openid-connect-discovery-1_0.html#IssuerDiscovery.")
	cmd.Flags().String("oidc.client-secret", "", "The OIDC client secret, see https://tools.ietf.org/html/rfc6749#section-2.3.")
	cmd.Flags().String("oidc.client-id", "", "The OIDC client ID, see https://tools.ietf.org/html/rfc6749#section-2.3.")
	cmd.Flags().String("oidc.audience", "", "The audience for whom the access token is intended, see https://openid.net/specs/openid-connect-core-1_0.html#IDToken.")

	return cmd
}
