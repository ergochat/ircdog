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

	"github.com/ergochat/irc-go/ircreader"
)

const (
	InitialBufferSize = 1024
	MaxBufferSize     = 1024 * 1024
)

var (
	// ErrorDisconnected indicates that this socket is disconnected.
	ErrorDisconnected = errors.New("Socket is disconnected")
)

type IRCSocket interface {
	SendLine(string) error
	GetLine() (string, error)
	Disconnect()
	RemoteAddr() net.Addr
}

// Socket appropriately buffers IRC lines.
type Socket struct {
	connection net.Conn

	reader ircreader.Reader

	writeMutex sync.Mutex
	closeOnce  sync.Once
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

	return MakeSocket(conn), nil
}

// MakeSocket makes a socket from the given connection.
func MakeSocket(conn net.Conn) *Socket {
	result := &Socket{
		connection: conn,
	}
	result.reader.Initialize(conn, InitialBufferSize, MaxBufferSize)
	return result
}

// GetLine returns a single IRC line from the socket.
func (s *Socket) GetLine() (string, error) {
	lineBytes, err := s.reader.ReadLine()
	return strings.TrimRight(string(lineBytes), "\r\n"), err
}

// SendLine sends a single IRC line to the socket
func (s *Socket) SendLine(line string) error {
	out := make([]byte, len(line)+2)
	copy(out, line[:])
	copy(out[len(line):], "\r\n")

	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()
	_, err := s.connection.Write(out)
	return err
}

// Disconnect severs our connection to the server.
func (s *Socket) Disconnect() {
	s.closeOnce.Do(s.realDisconnect)
}

func (s *Socket) realDisconnect() {
	s.connection.Close()
}

func (s *Socket) RemoteAddr() net.Addr {
	return s.connection.RemoteAddr()
}
