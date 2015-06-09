package conn

func cmdClose(args commandArgs, c *Conn) {
	c.setState(stateAuthenticated)
	c.selectedMailbox = nil
	c.writeResponse(args.ID(), "OK CLOSE Completed")
}
