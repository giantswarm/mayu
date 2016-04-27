# Flags of the 'mayu' binary

Mayu provides the following command line flags. When doing `mayu -h` you
should see this.

```nohighlight
Manage your bare metal machines

Usage:
  mayu [flags]
  
Flags:
      --alsologtostderr value           log to standard error as well as files
      --cluster-directory string        Path to the cluster directory (default "cluster")
      --config string                   Path to the configuration file (default "/etc/mayu/config.yaml")
  -d, --debug                           Print debug output
      --dnsmasq string                  Path to dnsmasq binary (default "/usr/sbin/dnsmasq")
      --dnsmasq-template string         Dnsmasq config template (default "./templates/dnsmasq_template.conf")
      --etcd-quorum-size int            Quorum of the etcd cluster (default 3)
      --first-stage-script string       Install script to install CoreOS on disk in the first stage. (default "./templates/first_stage_script.sh")
      --http-bind-address string        HTTP address Mayu listens on (default "0.0.0.0")
      --http-port int                   HTTP port Mayu listens on (default 4080)
      --ignition-config string          Final ignition config file that is used to boot the machine (default "./templates/ignition/gs_install.yaml")
      --images-cache-dir string         Directory for CoreOS images (default "./images")
      --last-stage-cloudconfig string   Final cloudconfig that is used to boot the machine (default "./templates/last_stage_cloudconfig.yaml")
      --log_backtrace_at value          when logging hits line file:N, emit a stack trace (default :0)
      --log_dir value                   If non-empty, write log files in this directory
      --logtostderr value               log to standard error instead of files (default true)
      --no-git                          Disable git operations
      --no-tls                          Disable tls
      --show-templates                  Show the templates and quit
      --static-html-path string         Path to Mayus binaries (eg. mayuctl, infopusher) (default "./static_html")
      --stderrthreshold value           logs at or above this threshold go to stderr (default 2)
      --template-snippets string        Cloudconfig or Ignition template snippets (eg storage or network configuration) (default "./template_snippets/cloudconfig")
      --tftproot string                 Path to the tftproot (default "./tftproot")
      --tls-cert-file string            Path to tls certificate file
      --tls-key-file string             Path to tls key file
      --use-ignition                    Use ignition configuration setup
  -v, --v value                         log level for V logs
      --version                         Show the version of Mayu
      --vmodule value                   comma-separated list of pattern=N settings for file-filtered logging
      --yochu-path string               Path to Yochus assets (eg docker, etcd, rkt binaries) (default "./yochu")
```
