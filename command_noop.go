package imap_server

func cmdNoop(args commandArgs, c *Conn) {
	c.writeResponse(args.Id(), "OK NOOP Completed")
}
