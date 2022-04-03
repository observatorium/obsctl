package e2e

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/efficientgo/e2e"
	e2edb "github.com/efficientgo/e2e/db"
	"github.com/efficientgo/tools/core/pkg/testutil"
)

// Adapted from https://github.com/observatorium/api/blob/main/test/e2e/services.go.

const (
	apiImage              = "quay.io/observatorium/api:latest"
	upImage               = "quay.io/observatorium/up:master-2021-02-12-03ef2f2"
	thanosImage           = "quay.io/thanos/thanos:v0.25.1"
	thanosRuleSyncerImage = "quay.io/observatorium/thanos-rule-syncer:main-2022-02-01-d4c24bc"
	rulesObjectStoreImage = "quay.io/observatorium/rules-objstore:main-2022-01-19-8650540"

	logLevelError = "error"
	// logLevelDebug = "debug"
)

type apiOptions struct {
	metricsReadEndpoint  string
	metricsWriteEndpoint string
	metricsRulesEndpoint string
}

type apiOption func(*apiOptions)

func withMetricsEndpoints(readEndpoint string, writeEndpoint string) apiOption {
	return func(o *apiOptions) {
		o.metricsReadEndpoint = readEndpoint
		o.metricsWriteEndpoint = writeEndpoint
	}
}

func withRulesEndpoint(rulesEndpoint string) apiOption {
	return func(o *apiOptions) {
		o.metricsRulesEndpoint = rulesEndpoint
	}
}

func newObservatoriumAPIService(
	e e2e.Environment,
	options ...apiOption,
) (e2e.InstrumentedRunnable, error) {
	opts := apiOptions{}
	for _, o := range options {
		o(&opts)
	}

	ports := map[string]int{
		"http":          8443,
		"http-internal": 8448,
	}

	args := e2e.BuildArgs(map[string]string{
		"--web.listen":           ":" + strconv.Itoa(ports["http"]),
		"--web.internal.listen":  ":" + strconv.Itoa(ports["http-internal"]),
		"--web.healthchecks.url": "http://127.0.0.1:8443",
		"--rbac.config":          filepath.Join("/shared/config", "rbac.yaml"),
		"--tenants.config":       filepath.Join("/shared/config", "tenants.yaml"),
		"--log.level":            logLevelError,
	})

	if opts.metricsReadEndpoint != "" && opts.metricsWriteEndpoint != "" {
		args = append(args, "--metrics.read.endpoint="+"http://"+opts.metricsReadEndpoint)
		args = append(args, "--metrics.write.endpoint="+"http://"+opts.metricsWriteEndpoint)
	}

	if opts.metricsRulesEndpoint != "" {
		args = append(args, "--metrics.rules.endpoint="+"http://"+opts.metricsRulesEndpoint)
	}

	return e2e.NewInstrumentedRunnable(e, "observatorium_api").WithPorts(ports, "http-internal").Init(
		e2e.StartOptions{
			Image:     apiImage,
			Command:   e2e.NewCommandWithoutEntrypoint("observatorium-api", args...),
			Readiness: e2e.NewHTTPReadinessProbe("http-internal", "/ready", 200, 200),
			User:      strconv.Itoa(os.Getuid()),
		},
	), nil
}

func newThanosReceiveService(e e2e.Environment) e2e.InstrumentedRunnable {
	ports := map[string]int{
		"http":         10902,
		"grpc":         10901,
		"remote_write": 19291,
	}

	args := e2e.BuildArgs(map[string]string{
		"--receive.local-endpoint": "0.0.0.0:" + strconv.Itoa(ports["grpc"]),
		"--label":                  "receive_replica=\"0\"",
		"--grpc-address":           "0.0.0.0:" + strconv.Itoa(ports["grpc"]),
		"--http-address":           "0.0.0.0:" + strconv.Itoa(ports["http"]),
		"--remote-write.address":   "0.0.0.0:" + strconv.Itoa(ports["remote_write"]),
		"--log.level":              logLevelError,
		"--tsdb.path":              "/tmp",
	})

	return e2e.NewInstrumentedRunnable(e, "thanos-receive").WithPorts(ports, "http").Init(
		e2e.StartOptions{
			Image:     thanosImage,
			Command:   e2e.NewCommand("receive", args...),
			Readiness: e2e.NewHTTPReadinessProbe("http", "/-/ready", 200, 200),
			User:      strconv.Itoa(os.Getuid()),
		},
	)
}

func newRulesObjstoreService(e e2e.Environment) e2e.InstrumentedRunnable {
	ports := map[string]int{"http": 8080, "internal": 8081}

	args := e2e.BuildArgs(map[string]string{
		"--log.level":            logLevelError,
		"--web.listen":           ":" + strconv.Itoa(ports["http"]),
		"--web.internal.listen":  ":" + strconv.Itoa(ports["internal"]),
		"--web.healthchecks.url": "http://127.0.0.1:" + strconv.Itoa(ports["http"]),
		"--objstore.config-file": filepath.Join("/shared/config", "rules-objstore.yaml"),
	})

	return e2e.NewInstrumentedRunnable(e, "rules_objstore").WithPorts(ports, "internal").Init(
		e2e.StartOptions{
			Image:     rulesObjectStoreImage,
			Command:   e2e.NewCommand("", args...),
			Readiness: e2e.NewHTTPReadinessProbe("internal", "/ready", 200, 200),
			User:      strconv.Itoa(os.Getuid()),
		},
	)
}

func newRuleSyncerService(e e2e.Environment, ruler string, rulesObjstore string) e2e.InstrumentedRunnable {
	ports := map[string]int{"http": 10911}
	args := e2e.BuildArgs(map[string]string{
		"--file":              filepath.Join("/shared/config", "rules.yaml"),
		"--rules-backend-url": "http://" + rulesObjstore,
		"--thanos-rule-url":   "http://" + ruler,
	})

	return e2e.NewInstrumentedRunnable(e, "rule_syncer").WithPorts(ports, "http").Init(
		e2e.StartOptions{
			Image:   thanosRuleSyncerImage,
			Command: e2e.NewCommand("", args...),
			User:    strconv.Itoa(os.Getuid()),
		},
	)
}

func newThanosRulerService(e e2e.Environment, query string) e2e.InstrumentedRunnable {
	ports := map[string]int{
		"http": 10904,
		"grpc": 10903,
	}

	args := e2e.BuildArgs(map[string]string{
		"--label":        "rule_replica=\"0\"",
		"--grpc-address": "0.0.0.0:" + strconv.Itoa(ports["grpc"]),
		"--http-address": "0.0.0.0:" + strconv.Itoa(ports["http"]),
		"--rule-file":    filepath.Join("/shared/config", "rules.yaml"),
		"--query":        query,
		"--log.level":    logLevelError,
		"--data-dir":     "/tmp",
	})

	return e2e.NewInstrumentedRunnable(e, "thanos-ruler").WithPorts(ports, "http").Init(
		e2e.StartOptions{
			Image:     thanosImage,
			Command:   e2e.NewCommand("rule", args...),
			Readiness: e2e.NewHTTPReadinessProbe("http", "/-/ready", 200, 200),
			User:      strconv.Itoa(os.Getuid()),
		},
	)
}

func startServicesForMetrics(t *testing.T, e e2e.Environment, envName string) (string, string, string) {
	thanosReceive := newThanosReceiveService(e)
	thanosRule := newThanosRulerService(e, "http://"+envName+"-"+"thanos-query:"+"9090")
	thanosQuery := e2edb.NewThanosQuerier(
		e,
		"thanos-query",
		[]string{thanosReceive.InternalEndpoint("grpc"), thanosRule.InternalEndpoint("grpc")},
		e2edb.WithImage(thanosImage),
	)

	testutil.Ok(t, e2e.StartAndWaitReady(thanosReceive, thanosQuery, thanosRule))

	bucket := "rulesobjstore"

	minio := e2edb.NewMinio(e, "rules-minio", bucket)
	testutil.Ok(t, e2e.StartAndWaitReady(minio))

	createRulesObjstoreYAML(t, e, bucket, minio.InternalEndpoint(e2edb.AccessPortName), e2edb.MinioAccessKey, e2edb.MinioSecretKey)

	rulesObjstore := newRulesObjstoreService(e)

	rulesSyncer := newRuleSyncerService(e, thanosRule.InternalEndpoint("http"), rulesObjstore.InternalEndpoint("http"))

	testutil.Ok(t, e2e.StartAndWaitReady(rulesObjstore))
	testutil.Ok(t, e2e.StartAndWaitReady(rulesSyncer))

	return thanosQuery.InternalEndpoint("http"),
		thanosReceive.InternalEndpoint("remote_write"),
		rulesObjstore.InternalEndpoint("http")
}

type runParams struct {
	initialDelay string
	period       string
	latency      string
	threshold    string
	duration     string
}

type upOptions struct {
	token     string
	runParams *runParams
}

type upOption func(*upOptions)

func withToken(token string) upOption {
	return func(o *upOptions) {
		o.token = token
	}
}

func withRunParameters(params *runParams) upOption {
	return func(o *upOptions) {
		o.runParams = params
	}
}

func newUpRun(
	env e2e.Environment,
	name string,
	readEndpoint, writeEndpoint string,
	options ...upOption,
) (e2e.InstrumentedRunnable, error) {
	opts := upOptions{}
	for _, o := range options {
		o(&opts)
	}

	ports := map[string]int{
		"http": 8888,
	}

	args := e2e.BuildArgs(map[string]string{
		"--listen":         "0.0.0.0:" + strconv.Itoa(ports["http"]),
		"--endpoint-type":  "metrics",
		"--endpoint-read":  readEndpoint,
		"--endpoint-write": writeEndpoint,
		"--log.level":      logLevelError,
		"--name":           "observatorium_write",
		"--labels":         "test=\"obsctl\"",
	})

	if opts.token != "" {
		args = append(args, "--token="+opts.token)
	}

	if opts.runParams != nil {
		if opts.runParams.initialDelay != "" {
			args = append(args, "--initial-query-delay="+opts.runParams.initialDelay)
		}
		if opts.runParams.duration != "" {
			args = append(args, "--duration="+opts.runParams.duration)
		}
		if opts.runParams.latency != "" {
			args = append(args, "--latency="+opts.runParams.latency)
		}
		if opts.runParams.threshold != "" {
			args = append(args, "--threshold="+opts.runParams.threshold)
		}
		if opts.runParams.period != "" {
			args = append(args, "--period="+opts.runParams.period)
		}
	}

	return e2e.NewInstrumentedRunnable(env, name).WithPorts(ports, "http").Init(
		e2e.StartOptions{
			Image:   upImage,
			Command: e2e.NewCommandWithoutEntrypoint("up", args...),
			User:    strconv.Itoa(os.Getuid()),
		},
	), nil
}
