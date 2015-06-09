package conn

func cmdLSub(args commandArgs, c *Conn) {
	for _, mailbox := range c.User.Mailboxes() {
		c.writeResponse("", "LSUB () \"/\" \""+mailbox.Name()+"\"")
	}
	c.writeResponse(args.ID(), "OK LSUB Completed")
}
