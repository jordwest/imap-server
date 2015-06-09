package conn

import "fmt"

func cmdSelect(args commandArgs, c *Conn) {
	var err error
	c.SelectedMailbox, err = c.User.MailboxByName(args.Arg(0))
	if err != nil {
		fmt.Fprintf(c, "%s NO %s\r\n", args.ID(), err)
		return
	}

	writeMailboxInfo(c, c.SelectedMailbox)
	c.writeResponse(args.ID(), "OK [READ-WRITE] SELECT completed")
}
