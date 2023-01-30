//go:build websocket

package lib

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

type IRCWebSocket struct {
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	closeOnce  sync.Once
	websocket  *websocket.Conn
}

func NewIRCWebSocket(wsUrl, origin string, tlsConfig *tls.Config) (IRCSocket, error) {
	var headers http.Header
	if origin != "" {
		headers = make(http.Header)
		u, err := url.Parse(origin)
		if err != nil {
			return nil, err
		}
		if u.Scheme == "" {
			u.Scheme = "https"
		}
		headers.Set("Origin", u.String())
	}

	dialer := websocket.Dialer{
		Subprotocols:    []string{"text.ircv3.net", "binary.ircv3.net"},
		TLSClientConfig: tlsConfig,
	}
	ws, resp, err := dialer.Dial(wsUrl, headers)
	if err != nil {
		return nil, fmt.Errorf("%w: %d", err, resp.StatusCode)
	}
	return &IRCWebSocket{
		websocket: ws,
	}, nil
}

func (w *IRCWebSocket) GetLine() (string, error) {
	w.readMutex.Lock()
	defer w.readMutex.Unlock()

	_, lineBytes, err := w.websocket.ReadMessage()
	return string(lineBytes), err
}

func (w *IRCWebSocket) SendLine(line string) error {
	messageType := websocket.TextMessage
	if w.websocket.Subprotocol() == "binary.ircv3.net" {
		messageType = websocket.BinaryMessage
	}
	w.writeMutex.Lock()
	defer w.writeMutex.Unlock()
	return w.websocket.WriteMessage(messageType, []byte(line))
}

func (w *IRCWebSocket) Disconnect() {
	w.closeOnce.Do(w.realDisconnect)
}

func (w *IRCWebSocket) realDisconnect() {
	w.websocket.Close()
}

func (w *IRCWebSocket) RemoteAddr() net.Addr {
	return w.websocket.RemoteAddr()
}
