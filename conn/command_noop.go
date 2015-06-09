package commands

func cmdNoop(args commandArgs, c *Conn) {
	c.writeResponse(args.ID(), "OK NOOP Completed")
}
