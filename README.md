# obsctl

A cli to interact with Observatorium instances.

```bash mdox-exec="obsctl --help"
usage: obsctl [<flags>] <command> [<args> ...]

obsctl

Flags:
  -h, --help               Show context-sensitive help (also try --help-long and
                           --help-man).
      --version            Show application version.
      --log.level=info     Log filtering level.
      --log.format=clilog  Log format to use.

Commands:
  help [<command>...]
    Show help.

  login [<flags>]
    Login as a tenant. Will also save tenant details locally.

  logout
    Logout currently logged in tenant.

  current
    Display configuration for the currently logged in tenant.

  switch [<tenant-name>]
    Switch to another locally saved tenant.

  read [<flags>]
    Read series, labels & rules of a tenant.

  rules [<flags>]
    Read/write Prometheus Rules configuration for a tenant.

  query <query>
    Query metrics for a tenant.


```
