package commands

const listArgSelector int = 1

func cmdList(args commandArgs, c *Conn) {
	if args.Arg(listArgSelector) == "" {
		// Blank selector means request directory separator
		c.writeResponse("", "LIST (\\Noselect) \"/\" \"\"")
	} else if args.Arg(listArgSelector) == "*" {
		// List all mailboxes requested
		for _, mailbox := range c.user.Mailboxes() {
			c.writeResponse("", "LIST () \"/\" \""+mailbox.Name()+"\"")
		}
	}
	c.writeResponse(args.ID(), "OK LIST completed")
}
