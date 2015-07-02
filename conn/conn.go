package conn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jordwest/imap-server/mailstore"
)

type connState int

const (
	StateNew connState = iota
	StateNotAuthenticated
	StateAuthenticated
	StateSelected
	StateLoggedOut
)

type WriteMode bool

const (
	ReadOnly  WriteMode = false
	ReadWrite           = true
)

const lineEnding string = "\r\n"

// Conn represents a client connection to the IMAP server
type Conn struct {
	state           connState
	Rwc             io.ReadWriteCloser
	Transcript      io.Writer
	recvReq         chan string
	Mailstore       mailstore.Mailstore // Pointer to the IMAP server's mailstore to which this connection belongs
	User            mailstore.User
	SelectedMailbox mailstore.Mailbox
	mailboxWritable WriteMode // True if write access is allowed to the currently selected mailbox
}

func NewConn(mailstore mailstore.Mailstore, netConn io.ReadWriteCloser, transcript io.Writer) (c *Conn) {
	c = new(Conn)
	c.Mailstore = mailstore
	c.Rwc = netConn
	c.Transcript = transcript
	return c
}

func (c *Conn) SetState(state connState) {
	c.state = state

	// As a precaution, reset any mailbox write access when changing states
	c.SetReadOnly()
}

func (c *Conn) SetReadOnly()  { c.mailboxWritable = ReadOnly }
func (c *Conn) SetReadWrite() { c.mailboxWritable = ReadWrite }

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

	return c.Rwc.Write(p)
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
	if c.state != StateNew {
		return errors.New("Welcome already sent")
	}
	c.writeResponse("", "OK IMAP4rev1 Service Ready")
	c.SetState(StateNotAuthenticated)
	return nil
}

func (c *Conn) assertAuthenticated(seq string) bool {
	if c.state != StateAuthenticated && c.state != StateSelected {
		c.writeResponse(seq, "BAD not authenticated")
		return false
	}

	if c.User == nil {
		panic("In authenticated state but no user is set")
	}

	return true
}

func (c *Conn) assertSelected(seq string, writable WriteMode) bool {
	// Ensure we are authenticated first
	if !c.assertAuthenticated(seq) {
		return false
	}

	if c.state != StateSelected {
		c.writeResponse(seq, "BAD not selected")
		return false
	}

	if c.SelectedMailbox == nil {
		panic("In selected state but no selected mailbox is set")
	}

	if writable == ReadWrite && c.mailboxWritable != ReadWrite {
		c.writeResponse(seq, "NO Selected mailbox is READONLY")
		return false
	}

	return true
}

// Close forces the server to close the client's connection
func (c *Conn) Close() error {
	fmt.Fprintf(c.Transcript, "Server closing connection\n")
	return c.Rwc.Close()
}

// Start tells the server to start communicating with the client (after
// the connection has been opened)
func (c *Conn) Start() error {
	if c.Rwc == nil {
		return errors.New("No connection exists")
	}

	c.recvReq = make(chan string)

	go func(ch chan string) {
		scanner := bufio.NewScanner(c.Rwc)
		for ok := scanner.Scan(); ok == true; ok = scanner.Scan() {
			text := scanner.Text()
			ch <- text
		}
		fmt.Fprintf(c.Transcript, "Client ended connection\n")
		close(ch)

	}(c.recvReq)

	for c.state != StateLoggedOut {
		// Always send welcome message if we are still in new connection state
		if c.state == StateNew {
			c.sendWelcome()
		}

		// Await requests from the client
		select {
		case req, ok := <-c.recvReq: // receive line of data from client
			if !ok {
				// The client has closed the connection
				c.state = StateLoggedOut
				break
			}
			fmt.Fprintf(c.Transcript, "C: %s\n", req)
			c.handleRequest(req)
		}
	}

	return nil
}
