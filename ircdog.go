// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/goshuirc/irc-go/ircmsg"

	docopt "github.com/docopt/docopt-go"
	"github.com/goshuirc/ircdog/lib"
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

Options:
	--tls               Connect using TLS.
	--tls-noverify      Don't verify the provided TLS certificates.
	-p --nopings        Don't automatically respond to incoming pings.
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
		TLS:  arguments["--tls"].(bool),
	}
	if arguments["--tls-noverify"].(bool) {
		connectionConfig.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// create new connection
	connection, err := lib.NewConnection(connectionConfig)
	if err != nil {
		log.Fatalf("Could not create new connection: %s\n", err.Error())
	}

	go func() {
		for {
			line, err := connection.GetLine()
			if err != nil {
				fmt.Println("Disconnected:", err.Error())
				connection.Disconnect()
				return
			}
			fmt.Println(line)

			// respond to incoming PINGs
			if !arguments["--nopings"].(bool) {
				msg, err := ircmsg.ParseLine(line)
				if err != nil {
					fmt.Println("** ircdog warning: this line looks incorrect **")
					continue
				}
				if msg.Command == "PING" {
					connection.SendMessage(true, nil, "", "PONG", msg.Params...)
				}
				//TODO(dan): Respond to CTCP PING/VERSION to make sure we don't get killed by nets
			}
		}
	}()

	// read incoming lines
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("** ircdog error: failed to read new input line:", err.Error())
			connection.Disconnect()
			return
		}

		err = connection.SendLine(strings.TrimRight(line, "\r\n"))
		if err != nil {
			fmt.Println("** ircdog error: failed to send line:", err.Error())
			connection.Disconnect()
			return
		}
	}
}
