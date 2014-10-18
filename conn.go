package imap_server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type ConnState int

const (
	StateNew ConnState = iota
	StateNotAuthenticated
	StateAuthenticated
	StateSelected
	StateLoggedOut
)

const lineEnding string = "\n"

// Represents a client connection to the IMAP server
type Conn struct {
	state      ConnState
	srv        *Server // Pointer to the IMAP server to which this connection belongs
	rwc        *net.Conn
	Transcript io.Writer
}

func (c *Conn) setState(state ConnState) {
	c.state = state
}

// Write a response to the client
func (c *Conn) writeResponse(seq string, command string) {
	if seq == "" {
		seq = "*"
	}
	if !strings.HasSuffix(command, lineEnding) {
		command += lineEnding
	}
	fmt.Fprintf(*c.rwc, "%s %s", seq, command)
	if c.Transcript != nil {
		fmt.Fprintf(c.Transcript, "S: %s %s", seq, command)
	}
}

func (c *Conn) sendWelcome() error {
	if c.state != StateNew {
		return errors.New("Welcome already sent")
	}
	c.writeResponse("", "OK IMAP4rev1 Service Ready")
	c.setState(StateNotAuthenticated)
	return nil
}

func (c *Conn) Start() error {
	if c.rwc == nil {
		return errors.New("No connection exists")
	}

	for {
		switch c.state {
		case StateNew:
			c.sendWelcome()
		}
	}

	return nil
}
