package e2e

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
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
			if resp.StatusCode == 409 {
				t.Log("Note: Hydra must be restarted for each run of the e2e test")
			}
			t.Fatalf("hydra register failed with status %d, body=%#v\n", resp.StatusCode, resp.Body)
		}
	}
}

func obtainToken(t *testing.T, realIssuerURL, identifiedIssuerUrl string, current int) string {
	ctx := oidc.InsecureIssuerURLContext(context.Background(), identifiedIssuerUrl)
	provider, err := oidc.NewProvider(ctx, "http://"+realIssuerURL+"/")
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

	if runtime.GOOS == "darwin" {
		ccc.TokenURL = strings.Replace(ccc.TokenURL, "host.docker.internal", "localhost", 1)
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

func notAssertResponse(t *testing.T, response string, expected string) {
	testutil.Assert(
		t,
		!strings.Contains(response, expected),
		fmt.Sprintf("failed to assert that the response '%s' contains '%s'", response, expected),
	)
}
