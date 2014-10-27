package imap_server

import (
	"encoding/base64"
	"fmt"
	"regexp"
)

// Register all supported client command handlers
// with the server. This function is run on server startup and
// panics if a command regex is invalid.
func (s *Server) registerCommands() {
	s.register("CAPABILITY", cmdCapability)
	s.register("LOGIN ([A-z0-9]+) ([A-z0-9]+)", cmdNA)
	s.register("AUTHENTICATE PLAIN", cmdAuthPlain)
	s.register("LIST", cmdList)
	s.register("LOGOUT", cmdLogout)
}

// Handles a CAPABILITY command
func cmdCapability(args commandArgs, c *Conn) {
	c.writeResponse("", "CAPABILITY IMAP4rev1 AUTH=PLAIN")
	c.writeResponse(args.Id(), "OK CAPABILITY completed")
}

// Handles a LOGIN command
func cmdLogin(args commandArgs, c *Conn) {
	c.writeResponse("", "BAD Not implemented")
}

// Handles PLAIN text AUTHENTICATE command
func cmdAuthPlain(args commandArgs, c *Conn) {
	// Compile login regex
	loginRE := regexp.MustCompile("(?:[A-z0-9]+)\x00([A-z0-9]+)\x00([A-z0-9]+)")

	// Tell client to go ahead
	c.writeResponse("+", "")

	// Wait for client to send auth details
	authDetails := <-c.recvReq
	data, err := base64.StdEncoding.DecodeString(authDetails)
	if err != nil {
		c.writeResponse("", "BAD Invalid auth details")
		return
	}
	fmt.Printf("Auth details received: %q\n", data)
	match := loginRE.FindSubmatch(data)
	fmt.Printf("Match: %q\n", match)
	if len(match) != 3 {
		c.writeResponse(args.Id(), "NO Incorrect username/password")
		return
	}
	c.setState(StateAuthenticated)
	c.writeResponse(args.Id(), "OK Authenticated")
}

func cmdList(args commandArgs, c *Conn) {
	c.writeResponse("", "LIST () \"/\" INBOX")
	c.writeResponse(args.Id(), "OK LIST Completed")
}

func cmdLogout(args commandArgs, c *Conn) {
	c.writeResponse("", "BYE IMAP4rev1 server logging out")
	c.setState(StateLoggedOut)
	c.writeResponse(args.Id(), "OK LOGOUT completed")
	c.Close()
}

func cmdNA(args commandArgs, c *Conn) {
	c.writeResponse("", "BAD Not implemented")
}
