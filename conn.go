package imap_server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type ConnState int

const (
	StateNew ConnState = iota
	StateNotAuthenticated
	StateAuthenticated
	StateSelected
	StateLoggedOut
)

const lineEnding string = "\r\n"

// Represents a client connection to the IMAP server
type Conn struct {
	state      ConnState
	srv        *Server // Pointer to the IMAP server to which this connection belongs
	rwc        net.Conn
	Transcript io.Writer
	recvReq    chan string
}

func (c *Conn) setState(state ConnState) {
	c.state = state
}

func (c *Conn) handleRequest(req string) {
	time.Sleep(2 * time.Second)
	for _, cmd := range c.srv.commands {
		matches := cmd.match.FindStringSubmatch(req)
		if len(matches) > 0 {
			cmd.handler(matches, c)
			return
		}
	}

	c.writeResponse("", "BAD Command not understood")
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
	fmt.Fprintf(c.rwc, "%s %s", seq, command)
	if c.Transcript != nil {
		fmt.Fprintf(c.Transcript, "S: %s %s", seq, command)
	}
}

// Send the server greeting to the client
func (c *Conn) sendWelcome() error {
	if c.state != StateNew {
		return errors.New("Welcome already sent")
	}
	c.writeResponse("", "OK IMAP4rev1 Service Ready")
	c.setState(StateNotAuthenticated)
	return nil
}

func (c *Conn) Close() error {
	fmt.Printf("Server closing connection\n")
	return c.rwc.Close()
}

func (c *Conn) Start() error {
	if c.rwc == nil {
		return errors.New("No connection exists")
	}

	c.recvReq = make(chan string)

	go func(ch chan string) {
		scanner := bufio.NewScanner(c.rwc)
		for ok := scanner.Scan(); ok == true; ok = scanner.Scan() {
			text := scanner.Text()
			fmt.Printf("Scanned %s\n", text)
			ch <- text
		}
		fmt.Printf("End of input?\n")
		close(ch)

	}(c.recvReq)

	for {
		// Always send welcome message if we are still in new connection state
		if c.state == StateNew {
			c.sendWelcome()
		}

		// Await requests from the client
		select {
		case req, ok := <-c.recvReq: // receive line of data from client
			if ok == false {
				return nil
			}
			if c.Transcript != nil {
				fmt.Fprintf(c.Transcript, "C: %s\n", req)
			}
			c.handleRequest(req)
		}
	}

	return nil
}
