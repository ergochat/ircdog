// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/goshuirc/irc-go/ircfmt"
	"github.com/goshuirc/irc-go/ircmsg"
	"github.com/goshuirc/irc-go/ircreader"

	docopt "github.com/docopt/docopt-go"
	"github.com/goshuirc/ircdog/lib"
	colorable "github.com/mattn/go-colorable"
)

func main() {
	version := lib.SemVer
	usage := `ircdog.
ircdog is a very simple telnet-like connection helper for IRC. Essentially, you
connect to an IRC server and send/see raw IRC protocol lines, which is very
useful for ircd and client developers.

What ircdog lets you do is ignore incoming PINGs (by automatically responding
to them) and see formatting characters (such as bold, colors and italics) on
incoming lines.

Usage:
	ircdog <host> <port> [options]
	ircdog -h | --help
	ircdog --version

Goshuirc Escapes:
	When the -r option is used, lines are displayed with the goshuirc escapes
	rather than as real formatted lines. goshuirc uses $ as an escape character
	along with these specific escapes:

	-------------------------------
	 Name          | Escape | Raw
	-------------------------------
	 Dollarsign    |   $$   |  $
	 Bold          |   $b   | 0x02
	 Colour        |   $c   | 0x03
	 Monospace     |   $m   | 0x11
	 Italic        |   $i   | 0x1d
	 Strikethrough |   $s   | 0x1e
	 Underscore    |   $u   | 0x1f
	 Reset         |   $r   | 0x0f
	-------------------------------

	Colours are followed by the specific colour code(s) in square brackets. For
	example, "$c[red,blue]" means red foreground, blue background. If there are
	no colour codes following, a pair of empty brackets like "$c[]" is used.

Sending Escapes:
	To send special characters like colour codes and CTCP messages, ircdog
	supports a few escape characters that get converted before messages are
	sent. These escapes are case-sensitive:

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
	---------------------------------

	These escapes are only enabled in standard mode (not listening mode),
	and can be disabled with the --no-controls option.

Options:
	--tls               Connect using TLS.
	--tls-noverify      Don't verify the provided TLS certificates.
	--listen=<address>  Listen on an address like ":7778", pass through traffic.
	--hide=<messages>   Comma-separated list of commands/numerics to not print.
	--no-italics        Don't use the ANSI italics code to represent italics.
	--no-controls       Don't use the control character escapes.
	-p --nopings        Don't automatically respond to incoming pings.
	-r --raw-incoming   Display incoming lines with raw goshuirc escapes.
	-h --help           Show this screen.
	--version           Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, version, false)

	host := arguments["<host>"].(string)
	portstring := arguments["<port>"].(string)
	port, err := strconv.Atoi(portstring)
	if err != nil || port < 1 || 65535 < port {
		log.Fatalln("Port must be a number 1-65535")
	}

	// create config
	connectionConfig := lib.ConnectionConfig{
		Host: host,
		Port: port,
		TLS:  arguments["--tls"].(bool) || arguments["--tls-noverify"].(bool),
	}
	if arguments["--tls-noverify"].(bool) {
		connectionConfig.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	colourablestdout := colorable.NewColorableStdout()

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

	// italics formatting code
	useItalics := !arguments["--no-italics"].(bool)

	// control code replacements
	useControlCodeReplacements := !arguments["--no-controls"].(bool)

	if arguments["--listen"] == nil {
		// not listening, just connect as usual
		// create new connection
		connection, err := lib.NewConnection(connectionConfig, &hiddenCommands)
		if err != nil {
			log.Fatalf("Could not create new connection: %s\n", err.Error())
		}

		go func() {
			for {
				line, err := connection.GetLine()
				if err != nil {
					fmt.Println("** ircdog disconnected:", err.Error())
					connection.Disconnect()
					os.Exit(0)
				}

				msg, err := ircmsg.ParseLine(line)
				if err != nil {
					fmt.Println("** ircdog warning: this next line looks incorrect, we're not formatting it **")
					fmt.Println(line)
					continue
				}

				// print line
				if !hiddenCommands[strings.ToUpper(msg.Command)] {
					if arguments["--raw-incoming"].(bool) {
						fmt.Println(ircfmt.Escape(line))
					} else {
						splitLine := lib.SplitLineIntoParts(line)
						fmt.Fprintln(colourablestdout, lib.AnsiFormatLineParts(splitLine, useItalics))
					}
				}

				// respond to incoming PINGs
				if !arguments["--nopings"].(bool) {
					if msg.Command == "PING" {
						connection.SendMessage(true, nil, "", "PONG", msg.Params...)
					}
					//TODO(dan): Respond to CTCP PING/VERSION to make sure we don't get killed by nets
				}
			}
		}()

		// read incoming lines
		var reader ircreader.Reader
		reader.Initialize(os.Stdin, lib.InitialBufferSize, lib.MaxBufferSize)
		for {
			lineBytes, err := reader.ReadLine()
			if err != nil {
				fmt.Println("** ircdog error: failed to read new input line:", err.Error())
				connection.Disconnect()
				return
			}
			line := string(lineBytes)

			if useControlCodeReplacements {
				line = lib.ReplaceControlCodes(line)
			}

			err = connection.SendLine(strings.TrimRight(line, "\r\n"))
			if err != nil {
				fmt.Println("** ircdog error: failed to send line:", err.Error())
				connection.Disconnect()
				return
			}
		}

	} else {
		// doing the listening dance, yay
		// use a mutext to make sure client and server don't talk over each other
		var outputMutex sync.Mutex

		listenAddress := arguments["--listen"].(string)

		ln, err := net.Listen("tcp", listenAddress)
		if err != nil {
			fmt.Println("** ircdog could not open listener:", err.Error())
			fmt.Println("Listener should have the form [host]:<port> like localhost:6667 or :8889")
			os.Exit(1)
		}

		fmt.Println("** ircdog listening on", listenAddress)
		fmt.Println("** ircdog will connect once we have a client connected on the listening port")

		// make the client connection
		clientConn, err := ln.Accept()
		if err != nil {
			fmt.Println("** ircdog could not accept incoming connection from listener:", err.Error())
			os.Exit(1)
		}

		client := lib.MakeSocket(clientConn)

		// create new server connection
		connection, err := lib.NewConnection(connectionConfig, &hiddenCommands)
		if err != nil {
			log.Fatalf("Could not create new connection: %s\n", err.Error())
		}

		go func() {
			colourablestdout := colorable.NewColorableStdout()
			for {
				line, err := connection.GetLine()
				if err != nil {
					fmt.Println("** ircdog server disconnected:", err.Error())
					client.Disconnect()
					connection.Disconnect()
					os.Exit(0)
				}

				msg, err := ircmsg.ParseLine(line)
				if err != nil {
					outputMutex.Lock()
					fmt.Println("** ircdog warning: this next line looks incorrect, we're not formatting it **")
					fmt.Println("<- ", line)
					outputMutex.Unlock()
					continue
				}

				// print line
				if !hiddenCommands[strings.ToUpper(msg.Command)] {
					if arguments["--raw-incoming"].(bool) {
						outputMutex.Lock()
						fmt.Println("<- ", ircfmt.Escape(line))
						outputMutex.Unlock()
					} else {
						splitLine := lib.AnsiFormatLineParts(lib.SplitLineIntoParts(line), useItalics)
						outputMutex.Lock()
						fmt.Fprintln(colourablestdout, "<-  "+splitLine)
						outputMutex.Unlock()
					}
				}

				err = client.SendLine(line)
				if err != nil {
					fmt.Println("** ircdog couldn't send line to client:", err.Error())
					client.Disconnect()
					connection.Disconnect()
					os.Exit(0)
				}
			}
		}()

		for {
			line, err := client.GetLine()
			if err != nil {
				fmt.Println("** ircdog client disconnected:", err.Error())
				client.Disconnect()
				connection.Disconnect()
				os.Exit(0)
			}

			msg, err := ircmsg.ParseLine(line)
			if err != nil {
				outputMutex.Lock()
				fmt.Println("** ircdog warning: this next line looks incorrect, we're not formatting it **")
				fmt.Println(" ->", line)
				outputMutex.Unlock()
				continue
			}

			// print line
			if !hiddenCommands[strings.ToUpper(msg.Command)] {
				if arguments["--raw-incoming"].(bool) {
					outputMutex.Lock()
					fmt.Println(" ->", ircfmt.Escape(line))
					outputMutex.Unlock()
				} else {
					outputMutex.Lock()
					splitLine := lib.AnsiFormatLineParts(lib.SplitLineIntoParts(line), useItalics)
					fmt.Fprintln(colourablestdout, " -> "+splitLine)
					outputMutex.Unlock()
				}
			}

			err = connection.SendLine(line)
			if err != nil {
				fmt.Println("** ircdog couldn't send line to server:", err.Error())
				client.Disconnect()
				connection.Disconnect()
				os.Exit(0)
			}
		}
	}
}
