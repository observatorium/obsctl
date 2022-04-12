package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/efficientgo/e2e"
	"github.com/efficientgo/tools/core/pkg/testutil"
	"github.com/observatorium/obsctl/pkg/cmd"
)

const (
	envName       = "obsctl-test"
	hydraURL      = "172.17.0.1:4444" // TODO(saswatamcode): Make this macOS-friendly.
	noOfTenants   = 2                 // Configure number of tenants.
	defaultTenant = 1                 // Set default tenant to use.
)

// preTest spins up all services required for metrics:
// - Receive
// - Query
// - Rule
// - Minio, Rules Objstore, Rule Syncer
// - Up
// Hydra is spun up externally via start_hydra.sh, as accessing it via docker network is difficult for obsctl.
// Follows similar pattern as https://observatorium.io/docs/usage/getting-started.md/.
// Also registers tenants in hydra.
func preTest(t *testing.T) *e2e.DockerEnvironment {
	e, err := e2e.NewDockerEnvironment(envName)
	testutil.Ok(t, err)
	t.Cleanup(e.Close)

	err = os.MkdirAll(filepath.Join(e.SharedDir(), "config"), 0750)
	testutil.Ok(t, err)

	registerHydraUsers(t, noOfTenants) // Only need to register this once.

	createTenantsYAML(t, e, hydraURL, noOfTenants)
	createRBACYAML(t, e, noOfTenants)

	read, write, rule := startServicesForMetrics(t, e, envName)

	api, err := newObservatoriumAPIService(e, withMetricsEndpoints(read, write), withRulesEndpoint(rule))
	testutil.Ok(t, err)
	testutil.Ok(t, e2e.StartAndWaitReady(api))
	testutil.Ok(t, os.MkdirAll(filepath.Join(e.SharedDir(), "obsctl"), 0750)) // Create config file beforehand.

	createObsctlConfigJson(t, e, hydraURL, "http://"+api.Endpoint("http")+"/", noOfTenants, defaultTenant)

	token := obtainToken(t, hydraURL, defaultTenant)

	up, err := newUpRun(
		e, "up-metrics-read-write",
		"http://"+api.InternalEndpoint("http")+"/api/metrics/v1/test-oidc-"+fmt.Sprint(defaultTenant)+"/api/v1/query",
		"http://"+api.InternalEndpoint("http")+"/api/metrics/v1/test-oidc-"+fmt.Sprint(defaultTenant)+"/api/v1/receive",
		withToken(token),
		withRunParameters(&runParams{period: "500ms", threshold: "1", latency: "10s", duration: "0"}),
	)

	createRulesYAML(t, e)

	testutil.Ok(t, e2e.StartAndWaitReady(up))
	testutil.Ok(t, err)

	time.Sleep(30 * time.Second) // Wait a bit for up to get some metrics in.

	return e
}

func TestObsctlMetricsCommands(t *testing.T) {
	e := preTest(t)

	t.Run("get labels for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.TODO(), filepath.Join(e.SharedDir(), "obsctl", "config.json"))

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "labels"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := ioutil.ReadAll(b)
		testutil.Ok(t, err)

		exp := `{
	"status": "success",
	"data": [
		"__name__",
		"receive_replica",
		"tenant_id",
		"test"
	]
}

`

		testutil.Equals(t, exp, string(got))
	})

	t.Run("get labelvalues for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.TODO(), filepath.Join(e.SharedDir(), "obsctl", "config.json"))

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "labelvalues", "--name=test"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := ioutil.ReadAll(b)
		testutil.Ok(t, err)

		exp := `{
	"status": "success",
	"data": [
		"obsctl"
	]
}

`

		testutil.Equals(t, exp, string(got))
	})

	t.Run("get rules for a tenant (none configured)", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.TODO(), filepath.Join(e.SharedDir(), "obsctl", "config.json"))

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := ioutil.ReadAll(b)
		testutil.Ok(t, err)

		exp := `{
	"status": "success",
	"data": {
		"groups": []
	}
}

`

		testutil.Equals(t, exp, string(got))
	})

	t.Run("get raw rules for a tenant (none configured)", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.TODO(), filepath.Join(e.SharedDir(), "obsctl", "config.json"))

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules.raw"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := ioutil.ReadAll(b)
		testutil.Ok(t, err)

		exp := `no rules found

`
		testutil.Equals(t, exp, string(got))
	})

	t.Run("set rules for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.TODO(), filepath.Join(e.SharedDir(), "obsctl", "config.json"))

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "set", "--rule.file=" + filepath.Join(e.SharedDir(), "obsctl", "rules.yaml")})
		testutil.Ok(t, contextCmd.Execute())

		got, err := ioutil.ReadAll(b)
		testutil.Ok(t, err)

		exp := "groups:\n- interval: 30s\n  name: test-firing-alert\n  rules:\n  - alert: TestFiringAlert\n    annotations:\n      description: Test firing alert\n      message: Message of firing alert here\n      summary: Summary of firing alert here\n    expr: vector(1)\n    for: 1m\n    labels:\n      severity: page\n\nsuccessfully updated rules file\n"

		testutil.Equals(t, exp, string(got))
	})

	t.Run("get rules.raw for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.TODO(), filepath.Join(e.SharedDir(), "obsctl", "config.json"))

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules.raw"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := ioutil.ReadAll(b)
		testutil.Ok(t, err)

		// Using assertResponse here as we cannot know exact tenant_id.
		assertResponse(t, string(got), "TestFiringAlert")
		assertResponse(t, string(got), "tenant_id")
	})

	t.Run("get rules for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.TODO(), filepath.Join(e.SharedDir(), "obsctl", "config.json"))

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules"})

		time.Sleep(30 * time.Second) // Wait a bit for rules to get synced.

		testutil.Ok(t, contextCmd.Execute())

		got, err := ioutil.ReadAll(b)
		testutil.Ok(t, err)

		// Using assertResponse here as we cannot know exact tenant_id.
		// As this is response from Query /api/v1/rules, should contain health data.
		assertResponse(t, string(got), "TestFiringAlert")
		assertResponse(t, string(got), "tenant_id")
		assertResponse(t, string(got), "health")

	})

	t.Run("get series for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.TODO(), filepath.Join(e.SharedDir(), "obsctl", "config.json"))

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "series", "--match", "observatorium_write"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := ioutil.ReadAll(b)
		testutil.Ok(t, err)

		// Using assertResponse here as we cannot know exact tenant_id.
		// As this is response from Query /api/v1/series, it should contain label of series written by up.
		assertResponse(t, string(got), "observatorium_write")
		assertResponse(t, string(got), "tenant_id")
		assertResponse(t, string(got), "test")
		assertResponse(t, string(got), "obsctl")
	})
}
