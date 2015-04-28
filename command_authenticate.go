package imap_server

import (
	"encoding/base64"
	"fmt"
	"regexp"
)

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
	if len(match) != 3 {
		c.writeResponse(args.Id(), "NO Incorrect username/password")
		return
	}
	c.user, err = c.srv.mailstore.Authenticate(string(match[1]), string(match[2]))
	if err != nil {
		c.writeResponse(args.Id(), "NO Incorrect username/password")
		return
	}
	c.setState(StateAuthenticated)
	c.writeResponse(args.Id(), "OK Authenticated")
}
