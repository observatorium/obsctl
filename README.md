# obsctl

A cli to interact with Observatorium instances.

```bash mdox-exec="obsctl --help"
CLI to interact with Observatorium

Usage:
  obsctl [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  context     Manage context configuration.
  help        Help about any command
  login       Login as a tenant. Will also save tenant details locally.
  logout      Logout a tenant. Will remove locally saved details.
  metrics     Metrics based operations for Observatorium.

Flags:
  -h, --help                help for obsctl
      --log.format string   Log format to use. (default "clilog")
      --log.level string    Log filtering level. (default "info")
  -v, --version             version for obsctl

Use "obsctl [command] --help" for more information about a command.
```

## Metrics

```bash mdox-exec="obsctl metrics --help"
Metrics based operations for Observatorium.

Usage:
  obsctl metrics [command]

Available Commands:
  get         Read series, labels & rules (JSON/YAML) of a tenant.
  query       Query metrics for a tenant.
  set         Write Prometheus Rules configuration for a tenant.

Flags:
  -h, --help   help for metrics

Global Flags:
      --log.format string   Log format to use. (default "clilog")
      --log.level string    Log filtering level. (default "info")

Use "obsctl metrics [command] --help" for more information about a command.
```
