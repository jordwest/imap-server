package commands

func cmdLogout(args commandArgs, c *Conn) {
	c.writeResponse("", "BYE IMAP4rev1 server logging out")
	c.setState(stateLoggedOut)
	c.writeResponse(args.ID(), "OK LOGOUT completed")
	c.Close()
}
