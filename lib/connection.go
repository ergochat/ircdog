// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/ergochat/irc-go/ircmsg"
)

type ConnectionConfig struct {
	Host      string
	Port      int
	TLS       bool
	TLSConfig *tls.Config
	Origin    string
}

type Connection struct {
	config         ConnectionConfig
	socket         IRCSocket
	hiddenCommands map[string]bool
}

// NewConnection returns a new Connection.
func NewConnection(config ConnectionConfig, hiddenCommands map[string]bool) (*Connection, error) {
	var socket IRCSocket
	var err error

	isWebsocket := false
	if u, uErr := url.Parse(config.Host); uErr == nil && (u.Scheme == "ws" || u.Scheme == "wss") {
		isWebsocket = true
	}

	if !isWebsocket {
		socket, err = ConnectSocket(config.Host, config.Port, config.TLS, config.TLSConfig)
		if err != nil {
			return nil, err
		}
	} else {
		socket, err = NewIRCWebSocket(config.Host, config.Origin, config.TLSConfig)
		if err != nil {
			return nil, err
		}
	}

	return &Connection{
		config:         config,
		socket:         socket,
		hiddenCommands: hiddenCommands,
	}, nil
}

// GetLine returns a single line from our socket.
func (conn *Connection) GetLine() (line string, err error) {
	line, err = conn.socket.GetLine()

	//TODO(dan): post-process line for colours, etc

	return line, err
}

// SendMessage assembles and sends an IRC message to the socket.
func (conn *Connection) SendMessage(print bool, tags map[string]string, prefix string, command string, params ...string) error {
	message := ircmsg.MakeMessage(tags, prefix, command, params...)
	line, err := message.Line()
	if err != nil {
		return err
	}

	// TODO reevaluate this
	line = strings.TrimRight(line, "\r\n")
	if print && !conn.hiddenCommands[strings.ToUpper(command)] {
		fmt.Println(line)
	}
	conn.socket.SendLine(line)
	return nil
}

// SendLine sends an IRC line to the socket.
func (conn *Connection) SendLine(line string) error {
	return conn.socket.SendLine(line)
}

// Disconnect disconnects from the server.
func (conn *Connection) Disconnect() {
	conn.socket.Disconnect()
}
