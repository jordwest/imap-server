package imap_server

// Handles PLAIN text LOGIN command
func cmdLogin(args commandArgs, c *Conn) {
	user, err := c.srv.mailstore.Authenticate(args.Arg(0), args.Arg(1))
	c.user = user
	if err != nil {
		c.writeResponse(args.Id(), "NO Incorrect username/password")
		return
	}
	c.setState(StateAuthenticated)
	c.writeResponse(args.Id(), "OK Authenticated")
}
