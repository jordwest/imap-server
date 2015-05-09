package imap

import "fmt"

func cmdExamine(args commandArgs, c *Conn) {
	m, err := c.user.MailboxByName(args.Arg(0))
	if err != nil {
		fmt.Fprintf(c, "%s NO %s\r\n", args.Id(), err)
		return
	}

	writeMailboxInfo(c, m)
	c.writeResponse(args.Id(), "OK [READ-ONLY] EXAMINE completed")
}
