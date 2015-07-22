package imap

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"

	"github.com/jordwest/imap-server/conn"
	"github.com/jordwest/imap-server/mailstore"
)

const (
	// defaultAddress is the default address that the IMAP server should listen
	// on.
	defaultAddress = ":143"
)

// Server represents an IMAP server instance.
type Server struct {
	Addr       string
	listeners  []net.Listener
	Transcript io.Writer
	mailstore  mailstore.Mailstore
}

// NewServer initialises a new Server. Note that this does not start the server.
// You must called either Listen() followed by Serve() or call ListenAndServe()
func NewServer(store mailstore.Mailstore) *Server {
	s := &Server{
		Addr:       defaultAddress,
		mailstore:  store,
		Transcript: ioutil.Discard,
	}
	return s
}

// ListenAndServe is shorthand for calling Serve() with a listener listening
// on the default port.
func (s *Server) ListenAndServe() (err error) {
	fmt.Fprintf(s.Transcript, "Listening on %s\n", s.Addr)
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("Error listening: %s\n", err)
	}

	return s.Serve(ln)
}

// Serve starts the server and spawns new goroutines to handle each client
// connection as they come in. This function blocks.
func (s *Server) Serve(l net.Listener) error {
	fmt.Fprintf(s.Transcript, "Serving on %s\n", l.Addr().String())
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("Error accepting connection: %s\n", err)
		}

		fmt.Fprintf(s.Transcript, "Connection accepted\n")
		c, err := s.newConn(conn)
		if err != nil {
			return err
		}

		go c.Start()
	}
}

func (s *Server) newConn(netConn net.Conn) (c *conn.Conn, err error) {
	c = conn.NewConn(s.mailstore, netConn, s.Transcript)
	c.SetState(conn.StateNew)
	return c, nil
}

// NewTestConnection is for test facilitation.
// Creates a server and then dials the server, returning the connection,
// allowing test to inject state and wait for an expected response
// The connection must be started manually with `go conn.Start()`
// once desired state has been injected
func NewTestConnection(transcript io.Writer) (s *Server, clientConn *textproto.Conn, serverConn *conn.Conn, server *Server, err error) {
	mStore := mailstore.NewDummyMailstore()
	s = NewServer(mStore)
	s.Addr = ":10143"
	s.Transcript = transcript

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	c, err := net.Dial("tcp4", "localhost:10143")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	textc := textproto.NewConn(c)
	clientConn = textc

	conn, err := l.Accept()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	fmt.Fprintf(s.Transcript, "Client connected\n")
	serverConn, err = s.newConn(conn)

	return s, clientConn, serverConn, s, nil
}
