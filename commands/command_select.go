package commands

import "fmt"

func cmdSelect(args commandArgs, c *Conn) {
	var err error
	c.selectedMailbox, err = c.user.MailboxByName(args.Arg(0))
	if err != nil {
		fmt.Fprintf(c, "%s NO %s\r\n", args.ID(), err)
		return
	}

	writeMailboxInfo(c, c.selectedMailbox)
	c.writeResponse(args.ID(), "OK [READ-WRITE] SELECT completed")
}
