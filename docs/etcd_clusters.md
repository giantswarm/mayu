# etcd Discovery

Mayu contains an internal etcd discovery to setup and manage your etcd clusters. During startup you can configure if you like to use an external etcd discovery or Mayu itself.

Use external etcd discovery:

```
mayu --etcd-discovery=https://discovery.etcd.io --use-internal-etcd-discovery=false --etcd-quorum-size=3
```

Use Mayus discovery:

```
mayu --etcd-endpoint=http://localhost:2379 --etcd-quorum-size=3
```

Note: Mayu defaults to the internal discovery. The parameter `--etcd-discovery` must be empty and `--use-internal-etcd-discovery` defaults to true.

As you see you also need an etcd endpoint for the internal etcd discovery in Mayu. Mayu only implements the creation of discovery tokens. All other requests are proxied to etcd. The data of the etcd clusters is only stored in etcd. Mayu automatically creates a first token and uses this token as default for all machines.

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
mayuctl set <host-serial> etcdtoken <your token>
```

Note: The token is only the last part of the full discovery url. eg http://localhost:4080/etcd/<token>

*Important*: If you change the etcd token of a host you need to reinstall the machine to make it happen. To reinstall just set the host state back to `configured` and reboot the machine. But if etcd on that host was part of the members of the old cluster you need to remove it as a member. Otherwise the new node will be part of the old and the new cluster as the members on the old clusters will connect to the new node again.

Prepare a host to be reinstalled:

```
mayuctl set <host-serial> state configured
```
