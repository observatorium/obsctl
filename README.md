# obsctl

A cli to interact with Observatorium instances.

```bash mdox-exec="obsctl --help"
CLI to interact with Observatorium

Usage:
  obsctl [flags]
  obsctl [command]

Available Commands:
  current     Display configuration for the currently logged in tenant.
  help        Help about any command
  login       Login as a tenant. Will also save tenant details locally.
  logout      Logout currently logged in tenant.
  query       Query metrics for a tenant.
  read        Read series, labels & rules of a tenant.
  rules       Read/write Prometheus Rules configuration for a tenant.
  switch      Switch to another locally saved tenant.

Flags:
  -h, --help                help for obsctl
      --log.format string   Log format to use. (default "clilog")
      --log.level string    Log filtering level. (default "info")
      --version             version for obsctl

Use "obsctl [command] --help" for more information about a command.
```
