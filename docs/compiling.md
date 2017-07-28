# Compiling Mayu

In order to compile Mayu you need to have `golang`  installed.

Use following command to compile `mayu` binary:
```
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .
```
This will compile statically linked `mayu` binary.

## Building docker image

After `mayu` binary successfully compiled docker image can be built:
```
docker build -t mayu .
```

## Updating vendors

To update the vendored libraries used by Mayu's binaries you need to have [glide](https://github.com/Masterminds/glide) installed.

Updating the vendored libraries is done by running the following `glide` targets:

```nohighlight
$ glide up -v
```
