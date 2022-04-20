package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/efficientgo/tools/core/pkg/testutil"
	"golang.org/x/oauth2/clientcredentials"
)

func registerHydraUsers(t *testing.T, noOfTenants int) {
	dataTmpl := `{"audience": ["observatorium-%[1]s"], "client_id": "user-%[1]s", "client_secret": "secret", "grant_types": ["client_credentials"], "token_endpoint_auth_method": "client_secret_basic"}`

	for i := 0; i < noOfTenants; i++ {
		d := fmt.Sprintf(dataTmpl, fmt.Sprint(i), fmt.Sprint(i))
		b := bytes.NewBuffer([]byte(d))
		resp, err := http.Post("http://127.0.0.1:4445/clients", "application/json", b)
		testutil.Ok(t, err)

		if resp.StatusCode/100 != 2 {
			t.Fatal(resp.Body)
		}
	}
}

func obtainToken(t *testing.T, issuerURL string, current int) string {
	provider, err := oidc.NewProvider(context.Background(), "http://"+issuerURL+"/")
	testutil.Ok(t, err)

	ccc := clientcredentials.Config{
		ClientID:     "user-" + fmt.Sprint(current),
		ClientSecret: "secret",
		TokenURL:     provider.Endpoint().TokenURL,
		Scopes:       []string{"openid", "offline_access"},
	}

	ccc.EndpointParams = url.Values{
		"audience": []string{"observatorium-" + fmt.Sprint(current)},
	}

	ts := ccc.TokenSource(context.Background())

	tkn, err := ts.Token()
	testutil.Ok(t, err)
	return tkn.AccessToken
}

func assertResponse(t *testing.T, response string, expected string) {
	testutil.Assert(
		t,
		strings.Contains(response, expected),
		fmt.Sprintf("failed to assert that the response '%s' contains '%s'", response, expected),
	)
}
