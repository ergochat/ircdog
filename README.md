# ircdog

`ircdog` is a tool for connecting to IRC servers and sending and receiving raw IRC protocol lines, similar to telnet or netcat, but with additional features:

* Automatically responds to `PING`, keeping the connection alive without active user input (`-p` disables)
* Renders [IRC formatting codes](https://modern.ircdocs.horse/formatting.html) for terminal display (`--raw` disables)
* Supports connecting to servers over plaintext, TLS, or [WebSocket](https://ircv3.net/specs/extensions/websocket)
* Can run as an intercepting proxy between another client and the server
* Can produce a transcript of raw traffic
* Supports escape sequences to easily send arbitrary binary data (`--raw` disables)
* Supports TLS client certificates

ircdog is primarily intended for IRC protocol developers who need to debug client or server behavior.

For more details, see the online help: `ircdog --help`

---

[![Example](docs/example.gif)](https://asciinema.org/a/bqmBrV8aIWDhvQqaxfpJrtj7r)

## License

ircdog is licensed under the attached ISC license.
