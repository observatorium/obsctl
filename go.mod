module github.com/observatorium/obsctl

go 1.17

require (
	github.com/bwplotka/mdox v0.9.0
	github.com/coreos/go-oidc/v3 v3.1.0
	github.com/efficientgo/e2e v0.11.1
	github.com/efficientgo/tools/core v0.0.0-20210731122119-5d4a0645ce9a
	github.com/go-kit/log v0.2.0
	github.com/google/uuid v1.2.0
	github.com/observatorium/api v1.0.0
	github.com/oklog/run v1.1.0
	github.com/spf13/cobra v1.4.0
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.9.1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/goleak v1.1.12 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20211111083644-e5c967477495 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/observatorium/api v1.0.0 => github.com/saswatamcode/api v0.1.3-0.20220331144432-9ccd3e239857 // TODO(saswatamcode): Remove once https://github.com/observatorium/api/pull/270 is merged.
