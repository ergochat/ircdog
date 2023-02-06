// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	docopt "github.com/docopt/docopt-go"
	supportscolor "github.com/jwalton/go-supportscolor"

	"github.com/ergochat/irc-go/ircfmt"
	"github.com/ergochat/irc-go/ircmsg"

	"github.com/ergochat/ircdog/lib"
)

// set via linker flags, either by make or by goreleaser:
// TODO these are not used yet (we're just using lib.SemVer)
var commit = ""  // git hash
var version = "" // tagged version

const (
	usage = `ircdog is a very simple telnet-like connection helper for IRC. It connects
to an IRC server and allows you to send and receive raw IRC protocol lines.
By default, ircdog will respond to incoming PING messages from the server,
keeping your connection alive without the need for active user input. It will
also render IRC formatting codes (such as boldface or color codes) for
terminal display.

Usage:
	ircdog <host> [<port>] [options]
	ircdog -h | --help
	ircdog --version

Sending Escapes:
	ircdog supports escape sequences in its input (use --raw to disable this).
	The following are case-sensitive:

	---------------------------------
	 Name          | Escape   | Raw
	---------------------------------
	 CTCP Escape   | [[CTCP]] | 0x01
	 Bold          | [[B]]    | 0x02
	 Colour        | [[C]]    | 0x03
	 Monospace     | [[M]]    | 0x11
	 Italic        | [[I]]    | 0x1d
	 Strikethrough | [[S]]    | 0x1e
	 Underscore    | [[U]]    | 0x1f
	 Reset         | [[R]]    | 0x0f
	 C hex escape  | [[\x??]] | 0x??
	---------------------------------

Options:
	--tls                 Connect using TLS.
	--tls-noverify        Don't verify the provided TLS certificates.
	--client-cert=<file>  A file containing a TLS client cert & key, to use for TLS connections.
	--listen=<address>    Listen on an address like ":7778", pass through traffic.
	--hide=<messages>     Comma-separated list of commands/numerics to not print.
	--origin=<url>        URL to send as the Origin header for a WebSocket connection
	-r --raw              Don't interpret IRC control codes when sending or receiving lines.
	--escape              Display incoming lines with irc-go escapes:
	                      https://pkg.go.dev/github.com/goshuirc/irc-go/ircfmt
	--italics             Enable ANSI italics codes (not widely supported).
	--color=<mode>        Override detected color support ('none', '16', '256')
	-p --nopings          Don't automatically respond to incoming pings.
	-h --help             Show this screen.
	--version             Show version.`
)

func parsePort(portStr string) (port int, err error) {
	if port, pErr := strconv.Atoi(portStr); pErr == nil && 1 <= port && port <= 65535 {
		return port, nil
	} else {
		return 0, fmt.Errorf("Invalid port number `%s`", portStr)
	}
}

func parseConnectionConfig(arguments map[string]any) (config lib.ConnectionConfig, err error) {
	tlsNoverify := arguments["--tls-noverify"].(bool)
	config.TLS = arguments["--tls"].(bool) || tlsNoverify

	host := arguments["<host>"].(string)

	u, uErr := url.Parse(host)
	if uErr != nil {
		err = fmt.Errorf("Invalid host: %w", uErr)
		return
	}
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else if u.Scheme == "http" {
		u.Scheme = "ws"
	}

	if u.Scheme == "" {
		// bare hostname, not a URL
		config.Host = strings.TrimPrefix(host, "unix:")
		portstring := arguments["<port>"]
		if portstring == nil {
			if config.TLS {
				config.Port = 6697
			} else if !strings.HasPrefix(host, "/") {
				err = fmt.Errorf("An explicit port number is required for plaintext (try 6667)")
				return
			}
		} else {
			config.Port, err = parsePort(portstring.(string))
			if err != nil {
				return
			}
		}
	} else if u.Scheme == "ws" || u.Scheme == "wss" {
		// WebsocketURL supersedes Host and Port options
		config.WebsocketURL = host
		if config.TLS && u.Scheme == "ws" {
			err = fmt.Errorf("To enable TLS on a WebSocket URL, use the scheme wss://")
			return
		}
	} else if u.Scheme == "irc" || u.Scheme == "ircs" {
		// ircs:// switches TLS on, but so should --tls with an irc:// URL
		if u.Scheme == "ircs" {
			config.TLS = true
		}
		if hostStr, portStr, splitErr := net.SplitHostPort(u.Host); splitErr == nil {
			config.Host = hostStr
			config.Port, err = parsePort(portStr)
			if err != nil {
				return
			}
		} else {
			config.Host = u.Host
			// no port in URL, use the protocol default
			if config.TLS {
				config.Port = 6697
			} else {
				config.Port = 6667
			}
		}
	}

	if originString := arguments["--origin"]; originString != nil {
		config.Origin = originString.(string)
	}

	if tlsNoverify {
		config.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	if clientCert := arguments["--client-cert"]; clientCert != nil {
		if config.TLSConfig == nil {
			config.TLSConfig = new(tls.Config)
		}

		clientCert, tErr := tls.LoadX509KeyPair(clientCert.(string), clientCert.(string))

		if tErr != nil {
			err = fmt.Errorf("Cannot load TLS client cert/key: %w", tErr)
			return
		}
		config.TLSConfig.Certificates = []tls.Certificate{clientCert}
	}

	return
}

func determineColorLevel(colorArg any) (colorLevel lib.ColorLevel) {
	// call this unconditionally for its side effects
	// (it does something to Windows terminals to make them ANSI-compliant)
	colorSupportResult := supportscolor.SupportsColor(os.Stdout.Fd(), supportscolor.SniffFlagsOption(false))
	colorLevel = lib.ColorLevel(colorSupportResult.Level)
	// now handle the override arg:
	if colorArg != nil {
		switch strings.ToLower(colorArg.(string)) {
		case "no", "none", "off", "false":
			colorLevel = lib.ColorLevelNone
		case "basic", "16", "ansi":
			colorLevel = lib.ColorLevelBasic
		case "256", "ansi256", "256color":
			colorLevel = lib.ColorLevelAnsi256
		case "16m", "ansi16m", "truecolor":
			// in practice this is treated the same as ColorLevelAnsi256
			colorLevel = lib.ColorLevelAnsi16m
		case "on", "yes":
			if colorLevel < lib.ColorLevelBasic {
				colorLevel = lib.ColorLevelBasic
			}
		case "default":
			// ok
		default:
			log.Fatalf("Invalid --color argument: `%s`", colorArg.(string))
		}
	}
	return
}

func main() {
	version := lib.SemVer
	arguments, _ := docopt.Parse(usage, nil, true, version, false)

	connectionConfig, err := parseConnectionConfig(arguments)
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}

	// list of commands/numerics to not print
	var hiddenString string
	if arguments["--hide"] != nil {
		hiddenString = arguments["--hide"].(string)
	}
	hiddenList := strings.Split(hiddenString, ",")
	hiddenCommands := make(map[string]bool)
	for _, cmd := range hiddenList {
		if 0 < len(cmd) {
			hiddenCommands[strings.ToUpper(cmd)] = true
		}
	}

	raw := arguments["--raw"].(bool)
	escape := arguments["--escape"].(bool)
	useItalics := arguments["--italics"].(bool)
	if raw && (escape || useItalics) {
		log.Fatalf("Cannot combine --raw with --escape or --italics")
	}
	answerPings := !arguments["--nopings"].(bool)

	colorLevel := determineColorLevel(arguments["--color"])

	var exitStatus int
	if listenAddr := arguments["--listen"]; listenAddr == nil {
		exitStatus = connectExternal(
			connectionConfig,
			hiddenCommands, raw, escape, answerPings, useItalics, colorLevel,
		)
	} else {
		exitStatus = listenAndConnectExternal(
			listenAddr.(string), connectionConfig,
			hiddenCommands, raw, escape, useItalics, colorLevel,
		)
	}
	os.Exit(exitStatus)
}

func connectExternal(
	connectionConfig lib.ConnectionConfig,
	hiddenCommands map[string]bool,
	raw, escape, answerPings, useItalics bool, colorLevel lib.ColorLevel) int {

	console, err := lib.NewStandardConsole()
	if err != nil {
		log.Printf("** ircdog could not initialize console: %s\n", err.Error())
		return 1
	}
	defer console.Close()

	connection, err := lib.NewConnection(connectionConfig)
	if err != nil {
		log.Printf("** ircdog could not create new connection: %s\n", err.Error())
		return 1
	}
	defer connection.Disconnect()

	// main goroutine will wait for either client or server EOF
	doneChan := make(chan struct{}, 2)
	done := func() {
		doneChan <- struct{}{}
	}

	// process incoming lines from server
	go func() {
		defer done()

		for {
			line, err := connection.GetLine()
			if err != nil {
				log.Println("** ircdog disconnected:", err.Error())
				return
			}

			msg, parseErr := ircmsg.ParseLine(line)

			if !(parseErr == nil && hiddenCommands[msg.Command]) {
				// print line
				if raw || parseErr != nil {
					fmt.Fprintln(console, line)
				} else if escape {
					fmt.Fprintln(console, ircfmt.Escape(line))
				} else {
					fmt.Fprintln(console, lib.IRCLineToAnsi(line, colorLevel, useItalics))
				}
			}

			// respond to incoming PINGs
			if parseErr == nil && answerPings && msg.Command == "PING" && len(msg.Params) != 0 {
				pong := makePong(msg)
				if !hiddenCommands["PONG"] {
					fmt.Fprintln(console, pong)
				}
				connection.SendLine(pong)
			}
		}
	}()

	// process incoming lines from user
	go func() {
		defer done()

		for {
			line, err := console.Readline()
			if err != nil {
				if err != io.EOF {
					log.Println("** ircdog error: failed to read new input line:", err.Error())
				}
				return
			}

			if !raw {
				line = lib.ReplaceControlCodes(line)
			}

			err = connection.SendLine(strings.TrimRight(line, "\r\n"))
			if err != nil {
				log.Println("** ircdog error: failed to send line:", err.Error())
				return
			}
		}
	}()

	<-doneChan
	return 0
}

func makePong(msg ircmsg.Message) string {
	// make a stylish irc-go PONG message that omits the : if possible
	// PONG parameter is the final parameter from PING:
	pongMsg := ircmsg.MakeMessage(nil, "", "PONG", msg.Params[len(msg.Params)-1])
	pong, _ := pongMsg.Line()
	pong = pong[:len(pong)-2] // trim \r\n
	return pong
}

type listenConnectionManager struct {
	ln               net.Listener
	connectionConfig lib.ConnectionConfig
	hiddenCommands   map[string]bool
	raw              bool
	escape           bool
	useItalics       bool
	colorLevel       lib.ColorLevel

	// prevent client and server from writing to stdout concurrently
	outputMutex sync.Mutex

	// allow at most one proxied connection at once:
	// 0 means no active connection, otherwise the unique ID of a connection
	activeConnection atomic.Uint64
}

func listenAndConnectExternal(
	listenAddress string, connectionConfig lib.ConnectionConfig,
	hiddenCommands map[string]bool,
	raw, escape, useItalics bool, colorLevel lib.ColorLevel) int {

	ln, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Println("** ircdog could not open listener:", err.Error())
		log.Println("Listener should have the form [host]:<port> like localhost:6667 or :8889")
		return 1
	}

	log.Printf("** ircdog listening on %s, waiting for client connection", listenAddress)

	manager := listenConnectionManager{
		ln:               ln,
		connectionConfig: connectionConfig,
		hiddenCommands:   hiddenCommands,
		raw:              raw,
		escape:           escape,
		useItalics:       useItalics,
		colorLevel:       colorLevel,
	}
	return manager.acceptLoop()
}

func (m *listenConnectionManager) acceptLoop() int {
	var connectionCounter uint64
	for {
		clientConn, err := m.ln.Accept()
		if err != nil {
			log.Printf("** ircdog could not accept incoming connection from listener: %v", err)
			return 1
		}
		connectionCounter++
		connectionID := connectionCounter
		if m.activeConnection.CompareAndSwap(0, connectionID) {
			log.Printf("** ircdog accepted connection from %s, connecting to remote", clientConn.RemoteAddr().String())
			// create new server connection
			server, err := lib.NewConnection(m.connectionConfig)
			if err != nil {
				log.Printf("** ircdog could not create new connection: %s\n", err.Error())
				clientConn.Write([]byte("ERROR :ircdog could not connect to remote server\r\n"))
				clientConn.Close()
				m.activeConnection.CompareAndSwap(connectionID, 0)
			}
			log.Printf("** ircdog connected to remote at %s", server.RemoteAddr().String())
			client := lib.MakeSocket(clientConn)
			go m.relay(connectionID, client, server, true)
			go m.relay(connectionID, server, client, false)
		} else {
			clientConn.Write([]byte("ERROR :ircdog already has an active connection\r\n"))
			clientConn.Close()
		}
	}
}

const (
	// printable indicators for whether the captured line is going from client to server,
	// or vice versa. note that markers are not shown at all in --raw mode:
	c2sMarkerPlain = " -> "
	s2cMarkerPlain = " <- "
	c2sMarkerColor = "\x1b[31;100m -> \x1b[0m"
	s2cMarkerColor = "\x1b[32;100m <- \x1b[0m"
)

func (m *listenConnectionManager) relay(connectionID uint64, input, output lib.IRCConnection, inputIsClient bool) {
	defer func() {
		input.Disconnect()
		output.Disconnect()
		m.activeConnection.CompareAndSwap(connectionID, 0)
	}()

	var inputName, outputName, marker string
	if inputIsClient {
		inputName, outputName, marker = "client", "server", c2sMarkerColor
		if m.escape || m.colorLevel == lib.ColorLevelNone {
			marker = c2sMarkerPlain
		}
	} else {
		inputName, outputName, marker = "server", "client", s2cMarkerColor
		if m.escape || m.colorLevel == lib.ColorLevelNone {
			marker = s2cMarkerPlain
		}
	}

	for {
		line, err := input.GetLine()
		if err != nil {
			log.Printf("** ircdog %s disconnected: %v", inputName, err.Error())
			return
		}

		msg, parseErr := ircmsg.ParseLine(line)
		if parseErr == nil && m.hiddenCommands[msg.Command] {
			continue
		}
		// print line
		m.outputMutex.Lock()
		if m.raw {
			fmt.Println(line)
		} else if m.escape {
			fmt.Printf("%s%s\n", marker, ircfmt.Escape(line))
		} else {
			fmt.Printf("%s%s\n", marker, lib.IRCLineToAnsi(line, m.colorLevel, m.useItalics))
		}
		m.outputMutex.Unlock()

		err = output.SendLine(line)
		if err != nil {
			log.Printf("** ircdog couldn't send line to %s: %v", outputName, err)
			return
		}
	}
}
