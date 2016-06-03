# Mayu with docker-compose

## Prerequisites

```bash
./scripts/fetch-coreos-image 1010.5.0

mkdir -p ./onsho/ipxe
curl https://github.com/giantswarm/onsho/raw/master/ipxe/ipxe.iso -o ./onsho/ipxe/ipxe.iso
```

## Running

```bash
make
docker-compose build

# start mayu
docker-compose up -d
docker-compose logs

# start vms
bridge_ifs="br-"$(docker network inspect monsho_default | jq -r '.[].Id[0:12]')
onsho create \
  --num-vms=3 \
  --bridge-ifs="$bridge_ifs" \
  --image=./onsho/ipxe/ipxe.iso

# watch vms via tmux
tmux a -t zoo

# list vms
mayu_ip_addr=$(docker inspect monsho_mayu_1 | jq -r '.[].NetworkSettings.Networks.monsho_default.IPAddress')
./mayuctl --host $mayu_ip_addr --no-tls list

# enter vm
ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null core@$ip_addr
```

## Cleanup

```bash
docker-compose down

onsho stop all ; onsho destroy all ; rm ~/.giantswarm/onsho -r
```
