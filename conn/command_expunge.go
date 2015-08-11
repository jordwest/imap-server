package conn

import (
	"fmt"
)

func cmdExpunge(args commandArgs, c *Conn) {
	if !c.assertSelected(args.ID(), readWrite) {
		return
	}

	// Delete flagged messages.
	msgs, err := c.SelectedMailbox.DeleteFlaggedMessages()
	if err != nil {
		c.writeResponse(args.ID(), "NO "+err.Error())
		return
	}

	// Write sequence numbers of deleted messages.
	for _, msg := range msgs {
		c.writeResponse("", fmt.Sprintf("%d EXPUNGE", msg.SequenceNumber()))
	}

	// And we're done.
	c.writeResponse(args.ID(), "OK EXPUNGE completed")
}
