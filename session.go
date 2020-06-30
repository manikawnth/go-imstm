package imstm

import (
	"crypto/tls"
	"net"
	"time"
)

// Session represents a client session for IMS connect connection.
type Session struct {
	// Addr is the IMS connect server address in the tcp address string format.
	// For example, "10.1.2.3:4567"
	Addr string

	// Datastore is the IMS datastore name.
	// If the length of the string is more than 8 bytes, only first 8 bytes are used
	DataStore string

	// ReadTimeout represents the client timeout for net.Conn read operations
	ReadTimeout time.Duration

	// WriteTimeout represents the client timeout for net.Conn write operations
	WriteTimeout time.Duration

	// TLSConfig is used for secure connections to IMS connect
	// If the value is nil, unsecure connection is established
	TLSConfig *tls.Config

	// tcp connection
	conn net.Conn
}

// Start returns a new connection to the IMS connect host
func (s *Session) Start() error {
	//validate the string
	var err error
	if _, err = net.ResolveTCPAddr("tcp", s.Addr); err != nil {
		return err
	}

	var dialer net.Dialer = net.Dialer{KeepAlive: -1}

	//dial the connection
	var conn net.Conn
	if s.TLSConfig != nil {
		conn, err = tls.DialWithDialer(&dialer, "tcp", s.Addr, s.TLSConfig)
	} else {
		conn, err = dialer.Dial("tcp", s.Addr)
	}
	s.conn = conn
	return err
}

// End ends the session
func (s *Session) End() error {
	return s.conn.Close()
}
