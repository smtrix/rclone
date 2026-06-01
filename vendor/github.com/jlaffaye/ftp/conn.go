package ftp

import (
	"crypto/tls"
	"net"
	"time"
)

// Options represents options that can be set on a ServerConn after login
type Options struct {
	// Charset is the character encoding to use for FTP commands and responses.
	// If empty, UTF-8 is used.
	Charset string
}

// DialWithTLSNoHandshake returns a DialOption that configures the ServerConn with specified TLS config
// but does not perform the TLS handshake immediately. The handshake will be performed on the first
// Read or Write. This is useful for FTP servers that hang during the TLS handshake.
func DialWithTLSNoHandshake(tlsConfig *tls.Config) DialOption {
	return DialOption{func(do *dialOptions) {
		do.tlsConfig = tlsConfig
		do.dialFunc = func(network, address string) (net.Conn, error) {
			conn, err := net.DialTimeout(network, address, do.dialer.Timeout)
			if err != nil {
				return nil, err
			}
			tlsConn := tls.Client(conn, tlsConfig)
			return tlsConn, nil
		}
	}}
}

// SetDeadline sets the read/write deadline on the underlying network connection.
func (c *ServerConn) SetDeadline(t time.Time) error {
	return c.netConn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline on the underlying network connection.
func (c *ServerConn) SetReadDeadline(t time.Time) error {
	return c.netConn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline on the underlying network connection.
func (c *ServerConn) SetWriteDeadline(t time.Time) error {
	return c.netConn.SetWriteDeadline(t)
}

// SetOptions sets options on the ServerConn after login.
// Currently only Charset is supported.
func (c *ServerConn) SetOptions(opts *Options) {
	// Charset handling would go here
	// For now this is a placeholder for compatibility
	_ = opts
}