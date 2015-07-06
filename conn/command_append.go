package conn

import (
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

	// Tell client to send the mail message
	c.writeResponse("+", "go ahead, feed me your message")

	// Read in the whole message
	messageData, err := c.ReadFixedLength(int(length))
	if err != nil {
		return
	}

	msg := mailbox.NewMessage()
	rawMsg, err := types.MessageFromBytes(messageData)
	if err != nil {
		c.writeResponse(args.ID(), "NO "+err.Error())
		return
	}
	msg = msg.SetHeaders(rawMsg.Headers)
	msg = msg.SetBody(rawMsg.Body)
	msg = msg.OverwriteFlags(flags)

	msg, err = msg.Save()
	if err != nil {
		c.writeResponse(args.ID(), "NO "+err.Error())
		return
	}

	c.writeResponse(args.ID(), "OK APPEND completed")
}
