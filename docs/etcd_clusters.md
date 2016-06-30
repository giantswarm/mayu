# Manage etcd clusters

Mayu contains an etcd discovery to setup and manage your etcd clusters. During startup you can configure if you like to use an external etcd discovery or Mayu itself.

Use external etcd discovery:

```
mayu --etcd-discovery=https://discovery.etcd.io --etcd-quorum-size=3
```

Use Mayus discovery:

```
mayu --etcd-endpoint=http://localhost:2379 --etcd-quorum-size=3
```

As you see you need to give Mayu access to an etcd endpoint. Mayu only implements the creation of discovery tokens. All other requests are proxied to etcd. The data of the etcd clusters is only stored in etcd. Mayu automatically creates a first token and uses this token as default for all machines. 

## Run etcd for the discovery

Start a single etcd instance in a container.

```
docker run --rm -v /usr/share/ca-certificates/:/etc/ssl/certs -p 4001:4001 -p 2380:2380 -p 2379:2379 \
  --name etcd quay.io/coreos/etcd \
  -name etcd0 \
  -advertise-client-urls http://127.0.0.1:2379,http://127.0.0.1:4001 \
  -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
  -initial-advertise-peer-urls http://127.0.0.1:2380 \
  -listen-peer-urls http://0.0.0.0:2380 \
  -initial-cluster-token etcd-cluster-1 \
  -initial-cluster etcd0=http://127.0.0.1:2380 \
  -initial-cluster-state new
```

## Show tokens in etcd

Mayu currently only allows to access a specific token. It is not implemented to list all tokens. To see all tokens you need to access etcd itself.

```
etcdctl ls --recursive /_etcd/registry
```

## Create a new etcd cluster token

To create a new token you need to send a `PUT` request to `/etcd/new`. The response will be a full URL to access and manage the etcd cluster. This is similar to how the official etcd discovery works.

```
curl -X PUT http://localhost:4080/etcd/new
```

## Change etcd cluster of a host

To change the etcd cluster of a host you need to overwrite the default etcd token and then reinstall the machine.

```
mayuctl override <host-serial> EtcdDiscoveryToken <your token>
```

Note: The token is only the last part of the full discovery url. eg http://localhost:4080/etcd/<token>
