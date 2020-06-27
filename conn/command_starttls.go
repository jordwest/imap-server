package conn

import (
	"bufio"
	"crypto/tls"
	"net"
)

func cmdStartTLS(args commandArgs, c *Conn) {
	if c.Secure {
		c.writeResponse(args.ID(), "BAD Already secure.")
		return
	}

	if c.StartTLSConfig == nil {
		c.writeResponse(args.ID(), "BAD STARTTLS not enabled.")
		return
	}

	RwcAsConn, ok := c.Rwc.(net.Conn)
	if !ok {
		c.writeResponse(args.ID(), "BAD Server-side problem: unexpected connection type.")
		return
	}

	c.writeResponse(args.ID(), "OK STARTTLS starting.")
	tlsServer := tls.Server(RwcAsConn, c.StartTLSConfig)
	c.Rwc = tlsServer
	c.RwcScanner = bufio.NewScanner(c.Rwc)
	c.Secure = true
}
