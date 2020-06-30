module github.com/giantswarm/mayu

go 1.14

require (
	github.com/beorn7/perks v1.0.1
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/giantswarm/mayu-infopusher v1.0.1
	github.com/giantswarm/microerror v0.0.0-20191011121515-e0ebc4ecf5a5
	github.com/giantswarm/micrologger v0.0.0-20191014091141-d866337f7393
	github.com/go-kit/kit v0.9.0
	github.com/go-logfmt/logfmt v0.5.0
	github.com/go-stack/stack v1.8.0
	github.com/golang/protobuf v1.3.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/json-iterator/go v1.1.9
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/client_model v0.1.0
	github.com/prometheus/common v0.7.0
	github.com/prometheus/procfs v0.0.8
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	go.etcd.io/etcd/v3 v3.3.0-rc.0.0.20200629155953-7f726db202a4 // indirect
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	gopkg.in/yaml.v2 v2.2.7
)

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
