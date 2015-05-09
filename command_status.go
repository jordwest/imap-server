package imap

import "fmt"

func cmdStatus(args commandArgs, c *Conn) {
	mailbox, err := c.user.MailboxByName(args.Arg(0))
	if err != nil {
		c.writeResponse(args.Id(), "NO "+err.Error())
		return
	}

	c.writeResponse("", fmt.Sprintf("STATUS %s (UIDNEXT %d UNSEEN %d)",
		mailbox.Name(), mailbox.NextUid(), mailbox.Unseen()))
	c.writeResponse(args.Id(), "OK STATUS Completed")
}
