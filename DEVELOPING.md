# Developing ircdog

Most development happens on the `develop` branch, which is occasionally rebased + merged into `master` when it's not incredibly broken. When this happens, the `develop` branch is usually pruned until I feel like making 'unsafe' changes again.

I may also name the branch `develop+feature` if I'm developing multiple, or particularly unstable, features.

The intent is to keep `master` relatively stable.


## Releasing a new version

1. Ensure dependencies are up-to-date.
1. Remove `-unreleased` from the version number in `lib/constants.go`.
1. Update the changelog with new changes.
1. Remove unused sections from the changelog, change the date/version number and write release notes.
1. Commit the new changelog and constants change.
1. Tag the release with `git tag --sign v0.0.0 -m "Release v0.0.0"` (`0.0.0` replaced with the real ver number).
1. Build binaries using `make release`
1. Sign the checksums file with `gpg --sign --detach-sig --local-user <fingerprint>`
1. Smoke-test a built binary locally
1. Point of no return: `git push origin master --tags` (this publishes the tag; any fixes after this will require a new point release)

Once it's built and released, you need to setup the new development version. To do so:

1. In `irc/constants.go`, update the version number to `0.0.1-unreleased`, where `0.0.1` is the previous release number with the minor field incremented by one (for instance, `0.9.2` -> `0.9.3-unreleased`).
2. At the top of the changelog, paste a new section with the content below.
3. Commit the new version number and changelog with the message `"Setup v0.0.1-unreleased devel ver"`.

**Unreleased changelog content**

```md
## Unreleased
New release of ircdog!

### Config Changes

### Security

### Added

### Changed

### Removed

### Fixed
```



## Updating `vendor/`

The `vendor/` directory holds our dependencies. When we import new repos, we need to update this folder to contain these new deps. This is something that I'll mostly be handling.

To update this folder:

1. Install https://github.com/golang/dep
2. `cd` to ircdog's folder
3. `dep ensure -update`
4. `cd vendor`
5. Commit the changes with the message `"Updated packages"`
6. `cd ..`
4. Commit the result with the message `"vendor: Updated submodules"`

This will make sure things stay nice and up-to-date for users.


## Debugging Hangs

To debug a hang, the best thing to do is to get a stack trace. Go's nice, and you can do so by running this:

    $ kill -ABRT <procid>

This will kill ircdog and print out a stack trace for you to take a look at.
