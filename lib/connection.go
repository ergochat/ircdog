// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"crypto/tls"
	"net"
)

type ConnectionConfig struct {
	// host-port pair for IRC over a normal stream transport
	Host string
	Port int
	// wss:// or ws:// URL for IRC over WebSocket
	WebsocketURL string
	TLS          bool
	TLSConfig    *tls.Config
	// Origin header for websockets
	Origin string
}

// IRCConnection is an abstract IRC connection.
type IRCConnection interface {
	// SendLine sends an IRC protocol line, given without \r\n
	SendLine(string) error
	// GetLine reads and returns an IRC protocol line, stripping the \r\n
	GetLine() (string, error)
	// Disconnect closes the connection, interrupting GetLine(); it must be
	// concurrency-safe and idempotent.
	Disconnect()
	RemoteAddr() net.Addr
}

func NewConnection(config ConnectionConfig) (conn IRCConnection, err error) {
	if config.WebsocketURL == "" {
		conn, err = ConnectSocket(config.Host, config.Port, config.TLS, config.TLSConfig)
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = NewIRCWebSocket(config.WebsocketURL, config.Origin, config.TLSConfig)
		if err != nil {
			return nil, err
		}
	}
	return
}
