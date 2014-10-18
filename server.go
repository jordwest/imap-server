package imap_server

import (
	"errors"
	"fmt"
	"net"
)

type Server struct {
	Addr     string
	listener net.Listener
}

func NewServer() *Server {
	return &Server{
		Addr: ":143",
	}
}

func (s *Server) ListenAndServe() (err error) {
	err = s.Listen()
	if err != nil {
		return err
	}
	return s.Serve()
}

func (s *Server) Listen() error {
	if s.listener != nil {
		return errors.New("Listener already exists")
	}
	fmt.Printf("Listening on %s\n", s.Addr)
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		fmt.Printf("Error listening: %s\n", err)
		return err
	}
	s.listener = ln
	return nil
}

// Starts the server and spawn new goroutines for each new client connection
func (s *Server) Serve() error {
	defer s.listener.Close()
	for {
		fmt.Printf("Waiting for connection...\n")
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Errorf("Error accepting connection: %s\n", err)
			return err
		} else {
			fmt.Printf("Connection accepted\n")
			c, err := s.newConn(&conn)
			if err != nil {
				return err
			}
			go c.Start()
		}
	}
}

func (s *Server) Close() (err error) {
	if s.listener == nil {
		return errors.New("Server not started")
	}
	err = s.listener.Close()
	if err == nil {
		s.listener = nil
	}
	return err
}

func (s *Server) newConn(conn *net.Conn) (c *Conn, err error) {
	c = new(Conn)
	c.srv = s
	c.rwc = conn
	c.setState(StateNew)
	return c, nil
}

// Test facilitation
// Creates a server and then dials the server, returning the connection,
// allowing test to inject state and wait for an expected response
// The connection must be started manually with `go conn.Start()`
// once desired state has been injected

func NewTestConnection() (s *Server, clientConn *net.Conn, serverConn *Conn, err error) {
	s = NewServer()
	s.Addr = ":10143"
	if err = s.Listen(); err != nil {
		return nil, nil, nil, err
	}

	if c, err := net.Dial("tcp4", "localhost:10143"); err != nil {
		return nil, nil, nil, err
	} else {
		clientConn = &c
	}

	conn, err := s.listener.Accept()
	if err != nil {
		return nil, nil, nil, err
	}
	serverConn, err = s.newConn(&conn)

	return s, clientConn, serverConn, nil
}
