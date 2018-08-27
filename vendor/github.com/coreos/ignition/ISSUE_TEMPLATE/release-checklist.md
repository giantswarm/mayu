---
name: Release checklist
about: Checklist for releasing a new version of Ignition
---

Release checklist:
 - [ ] Write release notes in NEWS. Get them reviewed and merged
     - [ ] If doing a branched release, also include a PR to merge the NEWS changes into master
 - [ ] Ensure your local copy is up to date with master and your working directory is clean
 - [ ] Ensure you can sign commits and any yubikeys/smartcards are plugged in
 - [ ] Run `./tag_release <vX.Y.z> <git commit hash>`
 - [ ] Push that tag to Github
 - [ ] Run `./build_releases`
 - [ ] Sign the release artifacts by running
```
gpg --local-user 0xCDDE268EBB729EC7 --detach-sign --armor <path to artifact>
```
for each release artifact. Do not try to sign all of them at once by globbing. If you do, gpg will sign the combination of all the release artifacts instead of each one individually.
 - [ ] Create a draft release on Github and upload all the release artifacts and their signatures. Copy and paste the release notes from NEWS here as well.
 - [ ] Publish the release
 - [ ] [Sync the docs](https://jenkins.prod.coreos.systems/job/coreos.com/job/sync-docs/build) using `ignition` for `PROJECT` and the version `X.Y.Z` (not `vX.Y.Z`) for `RELEASE`.
 - [ ] Bump the Ignition ebuild in [coreos-overlay](https://github.com/coreos/coreos-overlay)
 - [ ] Vendor the new Ignition version in [mantle](https://github.com/coreos/mantle)
