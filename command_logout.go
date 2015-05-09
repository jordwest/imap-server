package imap

func cmdLogout(args commandArgs, c *Conn) {
	c.writeResponse("", "BYE IMAP4rev1 server logging out")
	c.setState(StateLoggedOut)
	c.writeResponse(args.Id(), "OK LOGOUT completed")
	c.Close()
}
