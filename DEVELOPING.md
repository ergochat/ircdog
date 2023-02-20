# Developing ircdog

## Building from source

1. Obtain an [up-to-date distribution of the Go language for your OS and architecture](https://go.dev/dl)
1. Run `go version` to check that the expected version of Go is available on your `$PATH`
1. Check out your desired branch or tag and run `make`. This will produce an executable binary named `ircdog` in the base directory of the project. (All dependencies are vendored, so you will not need to fetch any dependencies remotely.)

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

## Debugging Hangs

To debug a hang, the best thing to do is to get a stack trace. Go's nice, and you can do so by running this:

    $ kill -ABRT <procid>

This will kill ircdog and print out a stack trace for you to take a look at.
