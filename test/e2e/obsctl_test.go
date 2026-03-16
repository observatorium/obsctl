package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/efficientgo/e2e"
	"github.com/efficientgo/tools/core/pkg/testutil"
	"github.com/observatorium/obsctl/pkg/cmd"
)

const (
	envName       = "obsctl-test"
	noOfTenants   = 2 // Configure number of tenants.
	defaultTenant = 1 // Set default tenant to use.
)

// preTest spins up all services required for metrics:
// - Receive
// - Query
// - Rule
// - Minio, Rules Objstore, Rule Syncer
// - Up
// - loki
// Hydra is spun up externally via start_hydra.sh, as accessing it via docker network is difficult for obsctl.
// Follows similar pattern as https://observatorium.io/docs/usage/getting-started.md/.
// Also registers tenants in hydra.

func preTest(t *testing.T) *e2e.DockerEnvironment {

	dir, err := os.Getwd()
	testutil.Ok(t, err)

	cmd := exec.Command("/bin/bash", dir+"/start_hydra.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	testutil.Ok(t, cmd.Run())

	e, err := e2e.NewDockerEnvironment(envName)
	testutil.Ok(t, err)
	t.Cleanup(e.Close)

	err = os.MkdirAll(filepath.Join(e.SharedDir(), "config"), 0750)
	testutil.Ok(t, err)

	hydraURL := "172.17.0.1:4444"
	switch runtime.GOOS {
	case "darwin":
		hydraURL = "docker.for.mac.localhost:4444"
	}

	registerHydraUsers(t, noOfTenants) // Only need to register this once.

	createTenantsYAML(t, e, hydraURL, noOfTenants)
	createRBACYAML(t, e, noOfTenants)

	bucket, endpoint, accessId, accessKey := startObjectStorageService(t, e)
	createLokiYAML(t, e, accessId, accessKey, endpoint, bucket)
	createRulesObjstoreYAML(t, e, bucket, endpoint, accessId, accessKey)

	read, write, rule := startServicesForMetrics(t, e, envName)
	logsEndpoint := startServicesForLogs(t, e)

	api, err := newObservatoriumAPIService(e, withMetricsEndpoints(read, write), withRulesEndpoint(rule), withLogsEndpoints(logsEndpoint))
	testutil.Ok(t, err)
	testutil.Ok(t, e2e.StartAndWaitReady(api))
	testutil.Ok(t, os.MkdirAll(filepath.Join(e.SharedDir(), "obsctl"), 0750)) // Create config file beforehand.

	createObsctlConfigJson(t, e, hydraURL, "http://"+api.Endpoint("http")+"/", noOfTenants, defaultTenant)

	token := obtainToken(t, hydraURL, defaultTenant)

	up, err := newUpRun(
		e, "up-metrics-read-write", "metrics",
		"http://"+api.InternalEndpoint("http")+"/api/metrics/v1/test-oidc-"+fmt.Sprint(defaultTenant)+"/api/v1/query",
		"http://"+api.InternalEndpoint("http")+"/api/metrics/v1/test-oidc-"+fmt.Sprint(defaultTenant)+"/api/v1/receive",
		withToken(token),
		withRunParameters(&runParams{period: "500ms", threshold: "1", latency: "10s", duration: "0"}),
	)

	createPrometheusRulesYAML(t, e)
	createLokiRulesYAML(t, e)

	testutil.Ok(t, e2e.StartAndWaitReady(up))
	testutil.Ok(t, err)

	up, err = newUpRun(
		e, "up-logs-read-write", "logs",
		"http://"+api.InternalEndpoint("http")+"/api/logs/v1/test-oidc-"+fmt.Sprint(defaultTenant)+"/loki/api/v1/query",
		"http://"+api.InternalEndpoint("http")+"/api/logs/v1/test-oidc-"+fmt.Sprint(defaultTenant)+"/loki/api/v1/push",
		withToken(token),
		withRunParameters(&runParams{period: "500ms", threshold: "1", latency: "10s", duration: "0"}),
	)

	testutil.Ok(t, err)
	testutil.Ok(t, e2e.StartAndWaitReady(up))

	time.Sleep(30 * time.Second) // Wait a bit for up to get some metrics in.

	return e

}

func TestObsctlMetricsCommands(t *testing.T) {

	e := preTest(t)
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(e.SharedDir(), "obsctl", "config.json")))

	t.Run("get ui", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "ui"})
		testutil.Ok(t, contextCmd.Execute())

		time.Sleep(time.Second * 500)

	})

	t.Run("get labels for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "labels"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
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

	t.Run("get labels for a tenant with match flag", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "labels", "--match=observatorium_write"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		// The response is the same with matcher too, as we have only one series with these exact labes
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

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "labelvalues", "--name=test"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
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

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
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

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules.raw"})
		testutil.NotOk(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		assertResponse(t, string(got), "no rules found")
	})

	t.Run("set rules for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "set", "--rule.file=" + filepath.Join(e.SharedDir(), "obsctl", "prometheus-rules.yaml")})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		exp := "successfully updated rules file\n"

		testutil.Equals(t, exp, string(got))
	})

	t.Run("get rules.raw for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules.raw"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		// Using assertResponse here as we cannot know exact tenant_id.
		assertResponse(t, string(got), "TestFiringAlert")
		assertResponse(t, string(got), "tenant_id")
	})

	t.Run("get rules for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules"})

		time.Sleep(30 * time.Second) // Wait a bit for rules to get synced.

		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		// Using assertResponse here as we cannot know exact tenant_id.
		// As this is response from Query /api/v1/rules, should contain health data.
		assertResponse(t, string(got), "TestFiringAlert")
		assertResponse(t, string(got), "tenant_id")
		assertResponse(t, string(got), "health")
	})

	t.Run("get rules for a tenant with type flag", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "rules", "--type=record"})

		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		notAssertResponse(t, string(got), "TestFiringAlert")
		notAssertResponse(t, string(got), "tenant_id")
		notAssertResponse(t, string(got), "health")
	})

	t.Run("get series for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "get", "series", "--match", "observatorium_write"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		// Using assertResponse here as we cannot know exact tenant_id.
		// As this is response from Query /api/v1/series, it should contain label of series written by up.
		assertResponse(t, string(got), "observatorium_write")
		assertResponse(t, string(got), "tenant_id")
		assertResponse(t, string(got), "test")
		assertResponse(t, string(got), "obsctl")
	})

	t.Run("query metrics for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"metrics", "query", "observatorium_write{test=\"obsctl\"}"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		assertResponse(t, string(got), "observatorium_write")
		assertResponse(t, string(got), "tenant_id")
		assertResponse(t, string(got), "test")
		assertResponse(t, string(got), "obsctl")
		assertResponse(t, string(got), "metric")
		assertResponse(t, string(got), "resultType")
		assertResponse(t, string(got), "vector")
	})

	t.Cleanup(func() {

		dir, err := os.Getwd()
		testutil.Ok(t, err)

		cmd := exec.Command("/bin/bash", dir+"/kill_hydra.sh")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		testutil.Ok(t, cmd.Run())

	})

}

func TestObsctlLogsCommands(t *testing.T) {

	e := preTest(t)
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(e.SharedDir(), "obsctl", "config.json")))

	t.Run("get labels for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "labels"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		exp := `{
	"status": "success",
	"data": [
		"__name__",
		"test"
	]
}

`

		testutil.Equals(t, exp, string(got))
	})

	t.Run("get labelvalues for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "labelvalues", "--name=test"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
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

	t.Run("get series for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "series", "--match", "observatorium_write"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		// Using assertResponse here as we cannot know exact tenant_id.
		// As this is response from Query /api/v1/series, it should contain label of series written by up.
		assertResponse(t, string(got), "observatorium_write")
		assertResponse(t, string(got), "tenant_id")
		assertResponse(t, string(got), "test")
		assertResponse(t, string(got), "obsctl")
		assertResponse(t, string(got), "receive_replica")

	})

	t.Run("query logs for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "query", "{test=\"obsctl\"}"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		assertResponse(t, string(got), "observatorium_write")
		assertResponse(t, string(got), "__name__")
		assertResponse(t, string(got), "log line 1")
		assertResponse(t, string(got), "test")
		assertResponse(t, string(got), "obsctl")
		assertResponse(t, string(got), "stream")
		assertResponse(t, string(got), "values")
		assertResponse(t, string(got), "resultType")
		assertResponse(t, string(got), "streams")
	})

	t.Run("get rules for a tenant (none configured)", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "rules"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		exp := `{
	"status": "success",
	"data": {
		"groups": []
	},
	"errorType": "",
	"error": ""
}
`

		testutil.Equals(t, exp, string(got))
	})

	t.Run("get raw rules for a tenant (none configured)", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "rules.raw"})
		err := contextCmd.Execute()
		testutil.NotOk(t, err)

		assertResponse(t, err.Error(), "no rule groups found")
	})

	t.Run("set rules for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "set", "--namespace", "logs", "--rule.file=" + filepath.Join(e.SharedDir(), "obsctl", "loki-rules.yaml")})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		exp := `{"status":"success","data":null,"errorType":"","error":""}
`

		testutil.Equals(t, exp, string(got))
	})

	t.Run("get all rules.raw namespaces for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "rules.raw"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		assertResponse(t, string(got), "TestFiringAlert")
	})

	t.Run("get all rules.raw groups in namespace for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "rules.raw", "--namespace", "logs"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		assertResponse(t, string(got), "TestFiringAlert")
	})

	t.Run("get a rules.raw group in namespace for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "rules.raw", "--namespace", "logs", "--group", "test-firing-alert"})
		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		assertResponse(t, string(got), "TestFiringAlert")
	})

	t.Run("get rules for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "rules"})

		time.Sleep(30 * time.Second) // Wait a bit for rules to get evaluated.

		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		assertResponse(t, string(got), "TestFiringAlert")
	})

	t.Run("get alerts for a tenant", func(t *testing.T) {
		b := bytes.NewBufferString("")

		contextCmd := cmd.NewObsctlCmd(context.Background())

		contextCmd.SetOut(b)
		contextCmd.SetArgs([]string{"logs", "get", "alerts"})

		time.Sleep(30 * time.Second) // Wait a bit for rules to get evaluated.

		testutil.Ok(t, contextCmd.Execute())

		got, err := io.ReadAll(b)
		testutil.Ok(t, err)

		assertResponse(t, string(got), "TestFiringAlert")
	})

	t.Cleanup(func() {
		dir, err := os.Getwd()
		testutil.Ok(t, err)

		cmd := exec.Command("/bin/bash", dir+"/kill_hydra.sh")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		testutil.Ok(t, cmd.Run())
	})
}
