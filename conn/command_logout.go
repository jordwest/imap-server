package conn

func cmdLogout(args commandArgs, c *Conn) {
	c.writeResponse("", "BYE IMAP4rev1 server logging out")
	c.SetState(StateLoggedOut)
	c.writeResponse(args.ID(), "OK LOGOUT completed")
	c.Close()
}
