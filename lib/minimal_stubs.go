//go:build minimal

package lib

import (
	"crypto/tls"
	"errors"
)

func NewIRCWebSocket(wsUrl, origin string, tlsConfig *tls.Config) (IRCConnection, error) {
	return nil, errors.New("websocket support disabled at compile time")
}
