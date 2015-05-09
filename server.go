package imap

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/textproto"
)

// Server represents an IMAP server instance
type Server struct {
	Addr       string
	listener   net.Listener
	commands   []command
	Transcript io.Writer
	mailstore  Mailstore
}

// NewServer initialises a new Server. Note that this does not start the server.
// You must called either Listen() followed by Serve() or call ListenAndServe()
func NewServer(store Mailstore) *Server {
	s := &Server{
		Addr:       ":143",
		commands:   make([]command, 0),
		mailstore:  store,
		Transcript: ioutil.Discard,
	}
	return s
}

// ListenAndServe is shorthand for calling Listen() followed by Serve().
func (s *Server) ListenAndServe() (err error) {
	err = s.Listen()
	if err != nil {
		return err
	}
	return s.Serve()
}

// Listen has the server begin listening for new connections.
// This function is non-blocking.
func (s *Server) Listen() error {
	if s.listener != nil {
		return errors.New("Listener already exists")
	}
	fmt.Fprintf(s.Transcript, "Listening on %s\n", s.Addr)
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		fmt.Printf("Error listening: %s\n", err)
		return err
	}
	s.listener = ln
	return nil
}

// Serve starts the server and spawns new goroutines to handle each client connection
// as they come in. This function blocks.
func (s *Server) Serve() error {
	defer s.listener.Close()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Errorf("Error accepting connection: %s\n", err)
			return err
		}

		fmt.Fprintf(s.Transcript, "Connection accepted\n")
		c, err := s.newConn(conn)
		if err != nil {
			return err
		}

		go c.Start()
	}
}

// Close stops the server listening for all new connections
func (s *Server) Close() (err error) {
	fmt.Fprintf(s.Transcript, "Closing server listener\n")
	if s.listener == nil {
		return errors.New("Server not started")
	}
	err = s.listener.Close()
	if err == nil {
		s.listener = nil
	}
	return err
}

func (s *Server) newConn(conn net.Conn) (c *Conn, err error) {
	c = new(Conn)
	c.srv = s
	c.rwc = conn
	c.setState(stateNew)
	c.Transcript = s.Transcript
	return c, nil
}

// NewTestConnection is for rest facilitation.
// Creates a server and then dials the server, returning the connection,
// allowing test to inject state and wait for an expected response
// The connection must be started manually with `go conn.Start()`
// once desired state has been injected
func NewTestConnection(transcript io.Writer) (s *Server, clientConn *textproto.Conn, serverConn *Conn, server *Server, err error) {
	mStore := NewDummyMailstore()
	s = NewServer(mStore)
	s.Addr = ":10143"
	s.Transcript = transcript
	if err = s.Listen(); err != nil {
		return nil, nil, nil, nil, err
	}

	c, err := net.Dial("tcp4", "localhost:10143")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	textc := textproto.NewConn(c)
	clientConn = textc

	conn, err := s.listener.Accept()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	fmt.Fprintf(s.Transcript, "Client connected\n")
	serverConn, err = s.newConn(conn)

	return s, clientConn, serverConn, s, nil
}
