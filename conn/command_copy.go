package conn

import (
	"strings"

	"github.com/jordwest/imap-server/mailstore"
	"github.com/jordwest/imap-server/types"
)

const (
	copyArgUID     int = 0
	copyArgRange   int = 1
	copyArgMailbox int = 2
)

func cmdCopy(args commandArgs, c *Conn) {
	if !c.assertSelected(args.ID(), readWrite) {
		return
	}

	// Check if the target mailbox exists.
	targetMailbox := args.Arg(copyArgMailbox)
	mbox, err := c.User.MailboxByName(targetMailbox)
	if err != nil {
		c.writeResponse(args.ID(), "NO [TRYCREATE] "+err.Error())
		return
	}

	// Check if connection is writable.
	if c.mailboxWritable != readWrite {
		c.writeResponse(args.ID(), "NO read-only connection")
		return
	}

	// Fetch the messages.
	seqSet, err := types.InterpretSequenceSet(args.Arg(copyArgRange))
	if err != nil {
		c.writeResponse(args.ID(), "NO "+err.Error())
		return
	}

	searchByUID := strings.ToUpper(args.Arg(copyArgUID)) == "UID "

	var msgs []mailstore.Message
	if searchByUID {
		msgs = c.SelectedMailbox.MessageSetByUID(seqSet)
	} else {
		msgs = c.SelectedMailbox.MessageSetBySequenceNumber(seqSet)
	}

	if len(msgs) == 0 {
		c.writeResponse(args.ID(), "NO no messages found")
		return
	}

	for _, msg := range msgs {
		_, err := mbox.NewMessage().
			SetBody(msg.Body()).
			SetHeaders(msg.Header()).
			AddFlags(msg.Flags() & types.FlagRecent).
			Save()

		if err != nil {
			// TODO Reverse all previous operations if it failed.
			c.writeResponse(args.ID(), "NO "+err.Error())
			return
		}
	}

	if searchByUID {
		c.writeResponse(args.ID(), "OK UID COPY Completed")
	} else {
		c.writeResponse(args.ID(), "OK COPY Completed")
	}
}
