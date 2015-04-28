package imap_server

const LIST_ARG_SELECTOR int = 1

func cmdList(args commandArgs, c *Conn) {
	if args.Arg(LIST_ARG_SELECTOR) == "" {
		// Blank selector means request directory separator
		c.writeResponse("", "LIST (\\Noselect) \"/\" \"\"")
	} else if args.Arg(LIST_ARG_SELECTOR) == "*" {
		// List all mailboxes requested
		for _, mailbox := range c.user.Mailboxes() {
			c.writeResponse("", "LIST () \"/\" \""+mailbox.Name()+"\"")
		}
	}
	c.writeResponse(args.Id(), "OK LIST completed")
}
