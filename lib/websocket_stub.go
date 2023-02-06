//go:build !websocket

package lib

import (
	"crypto/tls"
	"errors"
)

var (
	errNoWSSupport = errors.New("websocket support disabled at compile time")
)

func NewIRCWebSocket(wsUrl, origin string, tlsConfig *tls.Config) (IRCConnection, error) {
	return nil, errNoWSSupport
}
