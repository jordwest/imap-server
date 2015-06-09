package conn

func cmdClose(args commandArgs, c *Conn) {
	c.SetState(StateAuthenticated)
	c.SelectedMailbox = nil
	c.writeResponse(args.ID(), "OK CLOSE Completed")
}
