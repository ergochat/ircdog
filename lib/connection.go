// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"crypto/tls"
	"net/url"
)

type ConnectionConfig struct {
	Host      string
	Port      int
	TLS       bool
	TLSConfig *tls.Config
	Origin    string
}

func NewConnection(config ConnectionConfig) (IRCSocket, error) {
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

	return socket, err
}
