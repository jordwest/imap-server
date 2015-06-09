package conn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/jordwest/imap-server/mailstore"
)

type connState int

const (
	stateNew connState = iota
	stateNotAuthenticated
	stateAuthenticated
	stateSelected
	stateLoggedOut
)

const lineEnding string = "\r\n"

// Conn represents a client connection to the IMAP server
type Conn struct {
	state           connState
	srv             *Server // Pointer to the IMAP server to which this connection belongs
	rwc             net.Conn
	Transcript      io.Writer
	recvReq         chan string
	user            mailstore.User
	selectedMailbox mailstore.Mailbox
}

func (c *Conn) setState(state connState) {
	c.state = state
}

func (c *Conn) handleRequest(req string) {
	for _, cmd := range commands {
		matches := cmd.match.FindStringSubmatch(req)
		if len(matches) > 0 {
			cmd.handler(matches, c)
			return
		}
	}

	c.writeResponse("", "BAD Command not understood")
}

func (c *Conn) Write(p []byte) (n int, err error) {
	fmt.Fprintf(c.Transcript, "S: %s", p)

	return c.rwc.Write(p)
}

// Write a response to the client
func (c *Conn) writeResponse(seq string, command string) {
	if seq == "" {
		seq = "*"
	}
	// Ensure the command is terminated with a line ending
	if !strings.HasSuffix(command, lineEnding) {
		command += lineEnding
	}
	fmt.Fprintf(c, "%s %s", seq, command)
}

// Send the server greeting to the client
func (c *Conn) sendWelcome() error {
	if c.state != stateNew {
		return errors.New("Welcome already sent")
	}
	c.writeResponse("", "OK IMAP4rev1 Service Ready")
	c.setState(stateNotAuthenticated)
	return nil
}

// Close forces the server to close the client's connection
func (c *Conn) Close() error {
	fmt.Fprintf(c.Transcript, "Server closing connection\n")
	return c.rwc.Close()
}

// Start tells the server to start communicating with the client (after
// the connection has been opened)
func (c *Conn) Start() error {
	if c.rwc == nil {
		return errors.New("No connection exists")
	}

	c.recvReq = make(chan string)

	go func(ch chan string) {
		scanner := bufio.NewScanner(c.rwc)
		for ok := scanner.Scan(); ok == true; ok = scanner.Scan() {
			text := scanner.Text()
			ch <- text
		}
		fmt.Fprintf(c.Transcript, "Client ended connection\n")
		close(ch)

	}(c.recvReq)

	for c.state != stateLoggedOut {
		// Always send welcome message if we are still in new connection state
		if c.state == stateNew {
			c.sendWelcome()
		}

		// Await requests from the client
		select {
		case req, ok := <-c.recvReq: // receive line of data from client
			if !ok {
				// The client has closed the connection
				c.state = stateLoggedOut
				break
			}
			fmt.Fprintf(c.Transcript, "C: %s\n", req)
			c.handleRequest(req)
		}
	}

	return nil
}
