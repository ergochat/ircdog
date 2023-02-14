# Changelog
All notable changes to ircdog will be documented in this file.

This project adheres to [Semantic Versioning](http://semver.org/). For the purposes of versioning, we consider the "public API" to refer to the configuration files and CLI.

## [0.4.0] - 2023-02-19
ircdog v0.4.0 is a new release with fixes and enhancements:

* WebSocket support: pass a WebSocket URL, e.g. `wss://testnet.ergo.chat/webirc` as the host, omitting the separate port parameter
* `--origin` flag to set the `Origin` header for a WebSocket connection if necessary
* Support for `ircs://` and `irc://` URLs as the host, omitting a separate port parameter
* Support for TLS client certificates: pass `--client-cert=<file>` containing both the certificate and its private key in plaintext (#29, thanks [@jesopo](https://github.com/jesopo)!)
* Support for arbitrary C hex escapes in output, e.g. `[[\x00]]` to send the null byte (`--raw` disables interpretation of escapes)
* Experimental support for readline-like functionality, including up-arrow and Ctrl-R history retrieval. For now, this must be enabled explicitly with `--readline`
* `--transcript=<file>` appends a transcript of raw traffic to the specified file
* 256-color support (if supported by the terminal); use `--color=<none,16,256>` to override detected color support
* `--listen` supports reconnections without the need to restart ircdog (only one connection is allowed at a time)
* `--verbose` flag
* Fixed `--hide=PING` breaking automatic replies to `PING` (#23, thanks [@KoraggKnightWolf](https://github.com/KoraggKnightWolf)!)

## [0.3.0] - 2022-03-18
ircdog v0.3.0 is a new release with a few bug fixes:

* Upgrade [irc-go](https://github.com/ergochat/irc-go) to latest, fixing various correctness and performance issues
* Simplify switches; add `--raw` for raw I/O (i.e. control codes are not interpreted), `--tls-noverify` now entails `--tls` (#22)
* Support outgoing control code escapes

## [0.2.1] - 2017-12-29
Fix for a silly bug. Guess I should add some proper tests at some point!

### Fixed
* Fixed a locking bug that meant the `--listen` functionality was totally broken! Thanks [@jwheare](https://github.com/jwheare) for finding this bug!


## [0.2.0] - 2017-12-29
More formatting codes! Easier to see CTCP delimiters! Hiding messages and snooping on traffic!

All in all, this release includes a bunch of handy-to-have quality-of-life improvements. In particular, the new `--listen` functionality lets you sit in the middle of a client-server connection and see everything that goes on, and the new CTCP delimiter display should make things much clearer when you're debugging pesky CTCP messages.

### Added
* The CTCP delimiter `(0x01)` is now displayed very obviously as `[CTCP]` to help improve debugging issues around CTCP.
* We now support explicitly not using the ANSI italics code to represent italics with the `--no-italics` command-line arg (some terminals may not display it properly).
* We now support hiding commands/numerics with the `--hide` command-line arg. Thanks [@CarrotCodes](https://github.com/CarrotCodes) for the suggestion!
* We now support silently sitting in the middle of a (localhost) client and a server connection with the `--listen` command-line arg. Thanks [@DarkMio](https://github.com/DarkMio) for the suggestion!
* We now support the [reverse colour](https://modern.ircdocs.horse/formatting.html#reverse-color) formatting code `(0x07)`.


## [0.1.0] - 2017-12-26
Initial release of ircdog!

ircdog supports connecting, responding to pings, displaying formatting, and optionally displaying the raw decoded messages with goshuirc escapes.
