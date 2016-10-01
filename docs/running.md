# Running Mayu

## Download CoreOS Images

Before you start up Mayu you need to download a CoreOS image. In fact you need to download 
all the CoreOS versions that you specified in your `config.yaml`. You definitely need your
default CoreOS version. But you might also define different CoreOS versions within your profiles.

*Note:* the script is part of a release tarball or if you've build Mayu yourself it is in `bin-dist/`. Please do not run `scripts/fetch-coreos-image` - you need to fetch or build a distribution first.

```
./fetch-coreos-image 1122.2.0
```

If you like to distribute your own binaries for docker, etcd or fleet have a look at [Yochu](https://github.com/giantswarm/yochu).
There is also a script to fetch Giant Swarms binaries as an example.

```
./fetch-yochu-assets
```

## Start Mayu

Once Mayu is properly configured, it can be started:

```nohighlight
make bin-dist
./mayu --cluster-directory cluster -v=12 --no-git --no-tls
```

### Run Mayu within a Docker container

```
docker run --rm -it \
  --cap-add=NET_ADMIN \
  --net=host \
  --name=mayu \
  -v $(pwd)/cluster:/var/lib/mayu \
  -v $(pwd)/images:/usr/lib/mayu/images \
  -v $(pwd)/yochu:/usr/lib/mayu/yochu \
  giantswarm/mayu \
  -v=12 --no-git --no-tls
```

Or use the [`mayu.service`](https://github.com/giantswarm/mayu/blob/master/mayu.service) unit file included in this repository.

For running `mayu` in a local VM you might want to add two more volumes, to
enable DNS resultion by the `dnsmasq` included in `mayu`:

```
-v /etc/hosts:/etc/hosts -v /etc/resolv.conf:/etc/resolv.conf
```

## Cluster information

Mayu is now ready to bootstrap a new cluster. You can use [mayuctl](mayuctl.md) to list information about your cluster and machines.

Mayu uses the `cluster-directory` to save the cluster state:

```nohighlight
$ tree cluster
cluster
|-- 004b27ed-692e-b32e-1f68-d89aff66c71b
|   `-- conf.json
|-- 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a
|   `-- conf.json
|-- 7100c054-d2c9-e299-b669-e8bdb85f6904
|   `-- conf.json
|-- aa1f18e1-f14f-2dd9-4fa0-dae7317c712c
|   `-- conf.json
`-- cluster.json
```

Each cluster node has its own directory (identified by the serial number)
containing a JSON file with data about the node:

```json
{
  "Enabled": true,
  "Serial": "004b27ed-692e-b32e-1f68-d89aff66c71b",
  "MacAddresses": [
    "00:16:3e:a0:b7:df"
  ],
  "InternalAddr": "10.0.3.31",
  "Hostname": "00006811af601fe8",
  "MachineID": "00006811af601fe8e1d3f37902021ae0",
  "ConnectedNIC": "ens3",
  "FleetMetadata": [
    "rule-core=true"
  ],
  "LastBoot": "2015-10-08T19:14:36.227056826+02:00",
  "Profile": "core",
  "State": "running"
}
```

The cluster directory itself contains a `cluster.json` file with persistent
data about the cluster. If this file doesn't exist, it is initialized by
mayu.

```json
{
  "GitStore": true,
  "Config": {
    "EtcdDiscoveryURL": "https://discovery.etcd.io/e94768ef0f948b0c2e53536d9c5eeb8f"
  }
}
```

By default, mayu treats the cluster directory as a git repository, commiting
every change:

```nohighlight
$ git log --format="%ai => %s"
2015-10-08 19:14:37 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: updated state to running
2015-10-08 19:14:36 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: updated state to running
2015-10-08 19:14:35 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: updated state to running
2015-10-08 19:14:31 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: updated state to running
2015-10-08 19:13:28 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: updated state to installed
2015-10-08 19:13:28 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: updated state to installed
2015-10-08 19:13:28 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: updated state to installed
2015-10-08 19:13:28 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: updated state to installed
2015-10-08 19:10:54 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: updated host state to installing
2015-10-08 19:10:54 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: updated host connected nic
2015-10-08 19:10:54 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: updated host macAddress
2015-10-08 19:10:54 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: updated host profile and metadata
2015-10-08 19:10:54 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: updated host InternalAddr
2015-10-08 19:10:54 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: updated with predefined settings
2015-10-08 19:10:54 +0200 => aa1f18e1-f14f-2dd9-4fa0-dae7317c712c: host created
2015-10-08 19:10:54 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: updated host state to installing
2015-10-08 19:10:54 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: updated host connected nic
2015-10-08 19:10:54 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: updated host macAddress
2015-10-08 19:10:54 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: updated host state to installing
2015-10-08 19:10:54 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: updated host connected nic
2015-10-08 19:10:54 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: updated host macAddress
2015-10-08 19:10:54 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: updated host state to installing
2015-10-08 19:10:54 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: updated host connected nic
2015-10-08 19:10:54 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: updated host macAddress
2015-10-08 19:10:53 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: updated host profile and metadata
2015-10-08 19:10:53 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: updated host InternalAddr
2015-10-08 19:10:53 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: updated with predefined settings
2015-10-08 19:10:53 +0200 => 2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a: host created
2015-10-08 19:10:53 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: updated host profile and metadata
2015-10-08 19:10:53 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: updated host InternalAddr
2015-10-08 19:10:53 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: updated with predefined settings
2015-10-08 19:10:53 +0200 => 7100c054-d2c9-e299-b669-e8bdb85f6904: host created
2015-10-08 19:10:53 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: updated host profile and metadata
2015-10-08 19:10:53 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: updated host InternalAddr
2015-10-08 19:10:53 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: updated with predefined settings
2015-10-08 19:10:53 +0200 => 004b27ed-692e-b32e-1f68-d89aff66c71b: host created
2015-10-08 19:09:19 +0200 => generated etcd discovery url
2015-10-08 19:09:19 +0200 => initial commit
```
