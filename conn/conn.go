package conn

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jordwest/imap-server/mailstore"
)

type connState int

const (
	// StateNew is the initial state of a client connection; before a welcome
	// message is sent.
	StateNew connState = iota

	// StateNotAuthenticated is when a welcome message has been sent but hasn't
	// yet authenticated.
	StateNotAuthenticated

	// StateAuthenticated is when a client has successfully authenticated but
	// not yet selected a mailbox.
	StateAuthenticated

	// StateSelected is when a client has successfully selected a mailbox.
	StateSelected

	// StateLoggedOut is when a client has disconnected from the server.
	StateLoggedOut
)

type writeMode bool

const (
	readOnly  writeMode = false
	readWrite           = true
)

const lineEnding string = "\r\n"

// Conn represents a client connection to the IMAP server
type Conn struct {
	state           connState
	Rwc             io.ReadWriteCloser
	RwcScanner      *bufio.Scanner // Provides an interface for scanning lines from the connection
	Transcript      io.Writer
	Mailstore       mailstore.Mailstore // Pointer to the IMAP server's mailstore to which this connection belongs
	User            mailstore.User
	SelectedMailbox mailstore.Mailbox
	mailboxWritable writeMode // True if write access is allowed to the currently selected mailbox
	StartTLSConfig  *tls.Config //TLS configuration used for STARTTLS
	Secure          bool // True if the connection is secure
}

// NewConn creates a new client connection. It's intended to be directly used
// with a network connection. The transcript logs all client/server interactions
// and is very useful while debugging.
func NewConn(mailstore mailstore.Mailstore, netConn io.ReadWriteCloser, transcript io.Writer) (c *Conn) {
	c = new(Conn)
	c.Mailstore = mailstore
	c.Rwc = netConn
	c.Transcript = transcript
	return c
}

// SetState sets the state that an IMAP client is in. It also resets any mailbox
// write access.
func (c *Conn) SetState(state connState) {
	c.state = state

	// As a precaution, reset any mailbox write access when changing states
	c.SetReadOnly()
}

// SetReadOnly sets the client connection as read-only. It forbids any
// operations that may modify data.
func (c *Conn) SetReadOnly() { c.mailboxWritable = readOnly }

// SetReadWrite sets the connection as read-write.
func (c *Conn) SetReadWrite() { c.mailboxWritable = readWrite }

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

// Write a response to the client. Implements io.Writer.
func (c *Conn) Write(p []byte) (n int, err error) {
	fmt.Fprintf(c.Transcript, "S: %s", p)

	return c.Rwc.Write(p)
}

// Write a response to the client.
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

// Send the server greeting to the client.
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

func (c *Conn) assertSelected(seq string, writable writeMode) bool {
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

	if writable == readWrite && c.mailboxWritable != readWrite {
		c.writeResponse(seq, "NO Selected mailbox is READONLY")
		return false
	}

	return true
}

// Close forces the server to close the client's connection.
func (c *Conn) Close() error {
	fmt.Fprintf(c.Transcript, "Server closing connection\n")
	return c.Rwc.Close()
}

// ReadLine awaits a single line from the client.
func (c *Conn) ReadLine() (text string, ok bool) {
	ok = c.RwcScanner.Scan()
	return c.RwcScanner.Text(), ok
}

// ReadFixedLength reads data from the connection up to the specified length.
func (c *Conn) ReadFixedLength(length int) (data []byte, err error) {
	// Read the whole message into a buffer
	data = make([]byte, length)
	receivedLength := 0
	for receivedLength < length {
		bytesRead, err := c.Rwc.Read(data[receivedLength:])
		if err != nil {
			return data, err
		}
		receivedLength += bytesRead
	}

	return data, nil
}

// Start tells the server to start communicating with the client (after
// the connection has been opened).
func (c *Conn) Start() error {
	if c.Rwc == nil {
		return errors.New("No connection exists")
	}

	c.RwcScanner = bufio.NewScanner(c.Rwc)

	for c.state != StateLoggedOut {
		// Always send welcome message if we are still in new connection state
		if c.state == StateNew {
			c.sendWelcome()
		}

		// Await requests from the client
		req, ok := c.ReadLine()
		if !ok {
			// The client has closed the connection
			c.state = StateLoggedOut
			break
		}
		fmt.Fprintf(c.Transcript, "C: %s\n", req)
		c.handleRequest(req)

		if c.RwcScanner.Err() != nil {
			fmt.Fprintf(c.Transcript, "Scan error: %s\n", c.RwcScanner.Err())
		}
	}

	return nil
}
