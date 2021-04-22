// Copyright (c) 2017 Daniel Oaks <daniel@danieloaks.net>
// released under the ISC license

package lib

import (
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/goshuirc/irc-go/ircreader"
)

const (
	InitialBufferSize = 1024
	MaxBufferSize     = 1024 * 1024
)

var (
	// ErrorDisconnected indicates that this socket is disconnected.
	ErrorDisconnected = errors.New("Socket is disconnected")
)

// Socket appropriately buffers IRC lines.
type Socket struct {
	connection net.Conn

	connectedMutex sync.Mutex
	connected      bool

	readMutex sync.Mutex
	reader    ircreader.Reader

	writeMutex sync.Mutex
}

// ConnectSocket connects to the given host/port and starts our receivers if appropriate.
func ConnectSocket(host string, port int, useTLS bool, tlsConfig *tls.Config) (*Socket, error) {
	// assemble address
	address := net.JoinHostPort(host, strconv.Itoa(port))

	// initial connections
	var conn net.Conn
	var err error

	if useTLS {
		conn, err = tls.Dial("tcp", address, tlsConfig)
	} else {
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return nil, err
	}

	// set socket details
	s := Socket{
		connected:  true,
		connection: conn,
	}
	s.reader.Initialize(conn, InitialBufferSize, MaxBufferSize)

	return &s, nil
}

// MakeSocket makes a socket from the given connection.
func MakeSocket(conn net.Conn) *Socket {
	result := &Socket{
		connected:  true,
		connection: conn,
	}
	result.reader.Initialize(conn, InitialBufferSize, MaxBufferSize)
	return result
}

// GetLine returns a single IRC line from the socket.
func (s *Socket) GetLine() (string, error) {
	if !s.Connected() {
		return "", ErrorDisconnected
	}

	s.readMutex.Lock()
	defer s.readMutex.Unlock()
	lineBytes, err := s.reader.ReadLine()

	return strings.TrimRight(string(lineBytes), "\r\n"), err
}

// SendLine sends a single IRC line to the socket
func (s *Socket) SendLine(line string) error {
	if !s.Connected() {
		return ErrorDisconnected
	}

	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	out := make([]byte, len(line)+2)
	copy(out, line[:])
	copy(out[len(line):], "\r\n")
	_, err := s.connection.Write(out)
	return err
}

// Disconnect severs our connection to the server.
func (s *Socket) Disconnect() {
	s.connectedMutex.Lock()
	defer s.connectedMutex.Unlock()

	if !s.connected {
		s.connected = false
		s.connection.Close()
	}
}

// Connected returns true if we're still connected
func (s *Socket) Connected() bool {
	s.connectedMutex.Lock()
	defer s.connectedMutex.Unlock()

	return s.connected
}
