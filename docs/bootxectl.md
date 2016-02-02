# mayuctl

`mayuctl` is a client implementation to interact with mayu. By default TLS
is enabled. If your setup does not provide or rely on TLS for whatever reasons,
you can use the `--no-secure` flag to disable TLS and only communicate over
HTTP. The corresponding option for `mayu` can be set in the `config.yaml`.
There you just need to set `no_secure` to `true`.

### Help Usage

Lets have a look at the help usage.

```
Manage a mayu cluster

Usage:
  mayuctl [flags]
  mayuctl [command]

Available Commands:
  version     Show cli version
  list        List machines.
  status      Status of a host.

Flags:
  -d, --debug[=false]: Print debug output
      --host="localhost": Host name to connect to mayu service
      --no-secure[=false]: Do not use secure communication channels like https
      --port=4080: Port to connect to mayu service
  -v, --verbose[=false]: Print verbose output

Use "mayuctl [command] --help" for more information about a command.
```

### List Machines

Checking what machines are within a cluster can be done using the `list`
command. You should see output like this.

```
$ mayuctl list
IP           Serial                                Profile  State      Last Boot
172.17.8.31  004b27ed-692e-b32e-1f68-d89aff66c71b  core     "running"  2016-01-15 13:42:33.344687863 +0000 UTC
```

### Check Machine Status

To inspect a machine's status there is the `status` command. The output should
be something like this.

```
$ mayuctl status 004b27ed-692e-b32e-1f68-d89aff66c71b
Serial:              004b27ed-692e-b32e-1f68-d89aff66c71b
IP:                  172.17.8.31
IPMI:                <nil>
Macs:                08:00:27:6b:32:84
Cabinet:             0
Machine on Cabinet:  0
Hostname:            00007e267361d01f
MachineID:           00007e267361d01f233a3ed4900dcebb
ConnectedNIC:        enp0s3
Profile:             core
State:               "running"
Metadata:            role-core=true
LastBoot:            2016-01-15 13:42:33.344687863 +0000 UTC
Enabled:             true
```
