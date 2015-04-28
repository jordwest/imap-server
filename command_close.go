package imap_server

func cmdClose(args commandArgs, c *Conn) {
	c.setState(StateAuthenticated)
	c.selectedMailbox = nil
	c.writeResponse(args.Id(), "OK CLOSE Completed")
}
