# Compiling Mayu

In order to compile Mayu you need to have `golang`  installed.

To compile all binaries of the projects running a simple `go build` is sufficient.

## Updating vendors

To update the vendored libraries used by Mayu's binaries you need to havve [glide](https://github.com/Masterminds/glide) installed.

Updating the vendored libraries is done by running the following `glide` targets:

```nohighlight
$ glide up -v
```
