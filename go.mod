module github.com/giantswarm/mayu

go 1.14

require (
	github.com/coreos/bbolt v1.3.2 // indirect
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible // indirect
	github.com/giantswarm/mayu-infopusher v1.0.1
	github.com/giantswarm/microerror v0.0.0-20191011121515-e0ebc4ecf5a5
	github.com/giantswarm/micrologger v0.0.0-20191014091141-d866337f7393
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/juju/errgo v0.0.0-20140925100237-08cceb5d0b53 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d
	google.golang.org/grpc/examples v0.0.0-20211216223938-51835dc0912b // indirect
	gopkg.in/yaml.v2 v2.4.0
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.3
	github.com/coreos/etcd v3.3.15+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v3.2.1+incompatible
	github.com/gogo/protobuf v1.2.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v0.0.0-20170926233335-4201258b820c => github.com/gorilla/websocket v1.4.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
