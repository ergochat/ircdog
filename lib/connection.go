// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"crypto/tls"
	"errors"
	"strings"

	"github.com/goshuirc/irc-go/ircmsg"
)

type ConnectionConfig struct {
	Host      string
	Port      int
	TLS       bool
	TLSConfig *tls.Config
	Print     func(msg... string)
}

type Connection struct {
	config         ConnectionConfig
	socket         *Socket
	hiddenCommands *map[string]bool
}

// NewConnection returns a new Connection.
func NewConnection(config ConnectionConfig, hiddenCommands *map[string]bool) (*Connection, error) {
	if config.Port < 1 || 65535 < config.Port {
		return nil, errors.New("Port must be a number 1-65535")
	}

	socket, err := ConnectSocket(config.Host, config.Port, config.TLS, config.TLSConfig)
	if err != nil {
		return nil, err
	}

	return &Connection{
		config:         config,
		socket:         socket,
		hiddenCommands: hiddenCommands,
	}, nil
}

// GetLine returns a single line from our socket.
func (conn *Connection) GetLine() (string, error) {
	line, err := conn.socket.GetLine()

	//TODO(dan): post-process line for colours, etc

	return line, err
}

// SendMessage assembles and sends an IRC message to the socket.
func (conn *Connection) SendMessage(print bool, tags *map[string]ircmsg.TagValue, prefix string, command string, params ...string) error {
	message := ircmsg.MakeMessage(tags, prefix, command, params...)
	line, err := message.Line()
	if err != nil {
		return err
	}

	line = strings.TrimRight(line, "\r\n")
	if print && !(*conn.hiddenCommands)[strings.ToUpper(command)] {
		conn.config.Print(line)
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
