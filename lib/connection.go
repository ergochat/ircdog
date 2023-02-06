// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"crypto/tls"
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

func NewConnection(config ConnectionConfig) (IRCSocket, error) {
	var socket IRCSocket
	var err error

	if config.WebsocketURL == "" {
		socket, err = ConnectSocket(config.Host, config.Port, config.TLS, config.TLSConfig)
		if err != nil {
			return nil, err
		}
	} else {
		socket, err = NewIRCWebSocket(config.WebsocketURL, config.Origin, config.TLS, config.TLSConfig)
		if err != nil {
			return nil, err
		}
	}

	return socket, err
}
