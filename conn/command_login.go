package conn

// Handles PLAIN text LOGIN command
func cmdLogin(args commandArgs, c *Conn) {
	user, err := c.Mailstore.Authenticate(args.Arg(0), args.Arg(1))
	c.User = user
	if err != nil {
		c.writeResponse(args.ID(), "NO Incorrect username/password")
		return
	}
	c.SetState(StateAuthenticated)
	c.writeResponse(args.ID(), "OK Authenticated")
}
