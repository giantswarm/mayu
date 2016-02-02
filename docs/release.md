# Release A New Mayu Version

In general releases can be found here: https://github.com/giantswarm/mayu/releases

To prepare a release make sure all important changes are in master branch. Once
a git tag is created and pushed, checkout the git tag locally and create the
distribution package.

```
make bin-dist
```

This will create a `*.tar.gz` tarball you can upload to the github release
page. Because the tarball contains a lot of stuff (< 400 MB) the upload will
take a while. Make sure you provide a sufficient release title and desciption.
