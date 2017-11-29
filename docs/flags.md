# Flags of the 'mayu' binary

Mayu provides the following command line flags. When doing `mayu -h` you
should see this.

```nohighlight
Manage your bare metal machines

Usage:
  mayu [flags]

Flags:
      --alsologtostderr                  log to standard error as well as files
      --api-port int                     API HTTP port Mayu listens on (default 4080)
      --cluster-directory string         Path to the cluster directory (default "cluster")
      --config string                    Path to the configuration file (default "/etc/mayu/config.yaml")
  -d, --debug                            Print debug output
      --dnsmasq string                   Path to dnsmasq binary (default "/usr/sbin/dnsmasq")
      --dnsmasq-template string          Dnsmasq config template (default "./templates/dnsmasq_template.conf")
      --etcd-cafile string               The etcd CA file, if etcd is using non-trustred root CA certificate
      --etcd-discovery string            External etcd discovery base url (eg https://discovery.etcd.io). Note: This should be the base URL of the discovery without a specific token. Mayu itself creates a token for the etcd clusters.
      --etcd-endpoint string             The etcd endpoint for the internal discovery feature (you must also specify protocol). (default "http://127.0.0.1:2379")
      --etcd-quorum-size int             Default quorum of the etcd clusters (default 3)
      --files-dir string                 Directory for file templates (default "./files")
      --help                             Show mayu usage
      --http-bind-address string         HTTP address Mayu listens on (default "0.0.0.0")
      --ignition-config string           Final ignition config file that is used to boot the machine (default "./templates/ignition.yaml")
      --images-cache-dir string          Directory for Container Linux images (default "./images")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default true)
      --no-git                           Disable git operations
      --no-tls                           Disable tls
      --pxe-port int                     PXE HTTP port Mayu listens on (default 4081)
      --show-templates                   Show the templates and quit
      --static-html-path string          Path to Mayus binaries (eg. mayuctl, infopusher) (default "./static_html")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
      --template-snippets string         Cloudconfig or Ignition template snippets (eg storage or network configuration) (default "./templates/snippets/")
      --tftproot string                  Path to the tftproot (default "./tftproot")
      --tls-cert-file string             Path to tls certificate file
      --tls-key-file string              Path to tls key file
      --use-internal-etcd-discovery      Use the internal etcd discovery (default true)
  -v, --v Level                          log level for V logs
      --version                          Show the version of Mayu
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
      --yochu-path string                Path to Yochus assets (eg docker, etcd, rkt binaries) (default "./yochu")
```
