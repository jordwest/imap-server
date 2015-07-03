package conn

import (
	"fmt"
	"strconv"

	"github.com/jordwest/imap-server/types"
)

const (
	appendArgMailbox int = 0
	appendArgFlags   int = 1
	appendArgDate    int = 2
	appendArgLength  int = 3
)

// Add a new message to a mailbox
func cmdAppend(args commandArgs, c *Conn) {
	if !c.assertAuthenticated(args.ID()) {
		return
	}

	mailboxName := args.Arg(appendArgMailbox)
	mailbox, err := c.User.MailboxByName(mailboxName)
	if err != nil {
		c.writeResponse(args.ID(), "NO could not get mailbox")
		return
	}

	length, err := strconv.ParseUint(args.Arg(appendArgLength), 10, 64)
	if err != nil || length == 0 {
		c.writeResponse(args.ID(), "BAD invalid length for message literal")
		return
	}

	flagString := args.Arg(appendArgFlags)
	flags := types.Flags(0)
	if flagString != "" {
		flags = types.FlagsFromString(flagString)
	}

	fmt.Printf("Flags: %s. Length: %d. Mailbox: %s\n", flags.String(), length, mailbox.Name())

	// Tell client to send the mail message
	c.writeResponse("+", "go ahead, feed me your message")

	//msg := mailbox.NewMessage()
	// Length of data received so far
	receivedLength := uint64(0)
	// First receive headers
	body := false
	for receivedLength < length {
		line := <-c.recvReq
		// Blank line indicates end of headers, beginning of body
		if line == "" && body == false {
			body = true
			continue
		}

		if body == false {
			fmt.Printf("Received header: %s\n", line)
		} else {
			fmt.Printf("Received body: %s\n", line)
		}
	}

	c.writeResponse(args.ID(), "OK APPEND completed")
}
