# Changelog
All notable changes to ircdog will be documented in this file.

This project adheres to [Semantic Versioning](http://semver.org/). For the purposes of versioning, we consider the "public API" to refer to the configuration files and CLI.


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
