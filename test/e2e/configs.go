package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/efficientgo/e2e"
	"github.com/efficientgo/tools/core/pkg/testutil"
	"github.com/google/uuid"
	"github.com/observatorium/obsctl/pkg/config"
	"golang.org/x/oauth2"
)

const tenantsYAMLTmpl = `
- id: %[1]s
  name: %[2]s
  oidc:
    clientID: %[3]s
    issuerURL: %[4]s/`

func createTenantsYAML(
	t *testing.T,
	e e2e.Environment,
	issuerURL string,
	noOfTenants int,
) {

	var yamlContent []byte
	yamlContent = append(yamlContent, []byte(`tenants:`)...)

	for i := 0; i < noOfTenants; i++ {
		id := uuid.New()
		yamlContent = append(yamlContent, []byte(
			fmt.Sprintf(
				tenantsYAMLTmpl,
				id.String(),
				"test-oidc-"+fmt.Sprint(i),
				"observatorium-"+fmt.Sprint(i),
				"http://"+issuerURL),
		)...)

	}

	err := ioutil.WriteFile(
		filepath.Join(e.SharedDir(), "config", "tenants.yaml"),
		yamlContent,
		os.FileMode(0755),
	)

	testutil.Ok(t, err)
}

const rbacRoleBindingsYAML = `roleBindings:
- name: test
  roles:
  - read-write
  subjects:`

const rbacRoleBindingsYAMLTmpl = `
  - kind: user
    name: %[1]s`

const rbacRoleYAML = `
roles:
- name: read-write
  permissions:
  - read
  - write
  resources:
  - metrics
  tenants:`

const rbacRoleYAMLTmpl = `
  - %[1]s`

func createRBACYAML(
	t *testing.T,
	e e2e.Environment,
	noOfTenants int,
) {

	var yamlContent []byte
	yamlContent = append(yamlContent, []byte(rbacRoleBindingsYAML)...)

	for i := 0; i < noOfTenants; i++ {
		yamlContent = append(yamlContent, []byte(
			fmt.Sprintf(
				rbacRoleBindingsYAMLTmpl,
				"user-"+fmt.Sprint(i)),
		)...)
	}

	yamlContent = append(yamlContent, []byte(rbacRoleYAML)...)

	for i := 0; i < noOfTenants; i++ {
		yamlContent = append(yamlContent, []byte(
			fmt.Sprintf(
				rbacRoleYAMLTmpl,
				"test-oidc-"+fmt.Sprint(i)),
		)...)
	}

	err := ioutil.WriteFile(
		filepath.Join(e.SharedDir(), "config", "rbac.yaml"),
		yamlContent,
		os.FileMode(0755),
	)

	testutil.Ok(t, err)
}

const rulesObjstoreYAMLTpl = `
type: S3
config:
  bucket: %s
  endpoint: %s
  access_key: %s
  insecure: true
  secret_key: %s
`

func createRulesObjstoreYAML(
	t *testing.T,
	e e2e.Environment,
	bucket, endpoint, accessKey, secretKey string,
) {
	yamlContent := []byte(fmt.Sprintf(
		rulesObjstoreYAMLTpl,
		bucket,
		endpoint,
		accessKey,
		secretKey,
	))

	err := ioutil.WriteFile(
		filepath.Join(e.SharedDir(), "config", "rules-objstore.yaml"),
		yamlContent,
		os.FileMode(0755),
	)

	testutil.Ok(t, err)
}

func createObsctlConfigJson(
	t *testing.T,
	e e2e.Environment,
	issuerURL string,
	apiURL string,
	noOfTenants int,
	current int,
) {
	ctx := make(map[string]config.TenantConfig)

	for i := 0; i < noOfTenants; i++ {
		layout := "2006-01-02T15:04:05.000Z"
		str := "2022-03-20T16:49:34.000Z"
		ti, err := time.Parse(layout, str)
		testutil.Ok(t, err)

		ctx["test-oidc-"+fmt.Sprint(i)] = config.TenantConfig{
			Tenant: "test-oidc-" + fmt.Sprint(i),
			OIDC: &config.OIDCConfig{
				Audience:     "observatorium-" + fmt.Sprint(i),
				ClientID:     "user-" + fmt.Sprint(i),
				ClientSecret: "secret",
				IssuerURL:    "http://" + issuerURL + "/",
				Token: &oauth2.Token{
					Expiry:      ti,
					TokenType:   "bearer",
					AccessToken: "xyz",
				},
			},
		}
	}

	cfg := config.Config{
		APIs: map[string]config.APIConfig{
			"test-api": {
				URL:      apiURL,
				Contexts: ctx,
			},
		},
		Current: struct {
			API    string `json:"api"`
			Tenant string `json:"tenant"`
		}{
			API:    "test-api",
			Tenant: "test-oidc-" + fmt.Sprint(current),
		},
	}

	file, err := os.OpenFile(filepath.Join(e.SharedDir(), "obsctl", "config.json"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
	testutil.Ok(t, err)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	testutil.Ok(t, encoder.Encode(cfg))

	testutil.Ok(t, err)
}

const ruleYAML = `groups:
- interval: 30s
  name: test-firing-alert
  rules:
  - alert: TestFiringAlert
    annotations:
      description: Test firing alert
      message: Message of firing alert here
      summary: Summary of firing alert here
    expr: vector(1)
    for: 1m
    labels:
      severity: page
`

func createRulesYAML(
	t *testing.T,
	e e2e.Environment,
) {
	yamlContent := []byte(fmt.Sprint(
		ruleYAML,
	))

	err := ioutil.WriteFile(
		filepath.Join(e.SharedDir(), "obsctl", "rules.yaml"),
		yamlContent,
		os.FileMode(0755),
	)

	testutil.Ok(t, err)
}

const lokiYAML = `auth_enabled: true

server:
  http_listen_port: 3100

ingester:
  lifecycler:
    address: 0.0.0.0
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
    final_sleep: 0s
  chunk_idle_period: 5m
  chunk_retain_period: 30s

querier:
  engine:
    max_look_back_period: 5m
    timeout: 3m

schema_config:
  configs:
  - from: 2019-01-01
    store: boltdb
    object_store: filesystem
    schema: v11
    index:
      prefix: index_
      period: 168h

storage_config:
  boltdb:
    directory: /tmp/loki/index

  filesystem:
    directory: /tmp/loki/chunks

limits_config:
  enforce_metric_name: false
  reject_old_samples: false

`

func createLokiYAML(
	t *testing.T,
	e e2e.Environment,
) {
	yamlContent := []byte(fmt.Sprint(
		lokiYAML,
	))

	err := ioutil.WriteFile(
		filepath.Join(e.SharedDir(), "config", "loki.yml"),
		yamlContent,
		os.FileMode(0755),
	)

	testutil.Ok(t, err)
}
