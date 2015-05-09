package imap

func cmdClose(args commandArgs, c *Conn) {
	c.setState(stateAuthenticated)
	c.selectedMailbox = nil
	c.writeResponse(args.Id(), "OK CLOSE Completed")
}
