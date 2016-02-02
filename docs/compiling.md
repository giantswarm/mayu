# Compiling Mayu

In order to compile `mayu` you need to have `make` and [Docker](https://www.docker.com/) installed.

To compile all binaries of the projects running a simple `make` is sufficient.

## Updating vendors

To update the vendored libraries used by `mayu`'s binaries you need to have
`make` and [glide](https://github.com/Masterminds/glide) installed.

Updating the vendored libraries is done by running the following `make` targets:
```
make vendor-clean
make vendor-update
```
