package conn

// Handles a CAPABILITY command
func cmdCapability(args commandArgs, c *Conn) {
	if c.StartTLSConfig == nil {
		c.writeResponse("", "CAPABILITY IMAP4rev1 AUTH=PLAIN")
	} else {
		c.writeResponse("", "CAPABILITY IMAP4rev1 AUTH=PLAIN STARTTLS")
	}
	c.writeResponse(args.ID(), "OK CAPABILITY completed")
}
