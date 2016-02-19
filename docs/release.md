# Release A New Mayu Version

Releases can be found here: https://github.com/giantswarm/mayu/releases

To prepare a release, make sure all important changes are in the `master` branch. Once
a git tag is created and pushed, checkout the git tag locally and create the
distribution package.

```nohighlight
make bin-dist
```

This will create a `tar.gz` tarball you can upload to the GitHub release
page. Since the tarball contains a lot of stuff (< 400 MB), the upload will
take a while. Make sure you provide a sufficient release title and desciption.
