package conn

import (
	"fmt"
	"strings"

	"github.com/jordwest/imap-server/mailstore"
	"github.com/jordwest/imap-server/types"
)

const storeArgUID int = 0
const storeArgRange int = 1
const storeArgOperation int = 2
const storeArgSilent int = 3
const storeArgFlags int = 4

func cmdStoreFlags(args commandArgs, c *Conn) {
	operation := args.Arg(storeArgOperation)
	flags := args.Arg(storeArgFlags)
	uid := strings.ToUpper(args.Arg(storeArgUID)) == "UID "
	seqSetStr := args.Arg(storeArgRange)

	silent := false
	if args.Arg(storeArgSilent) == ".SILENT" {
		silent = true
	}
	if silent {
		fmt.Printf("Silently ")
	}

	var msgs []mailstore.Message
	seqSet, err := types.InterpretSequenceSet(seqSetStr)
	if err != nil {
		c.writeResponse(args.ID(), "NO "+err.Error())
		return
	}
	if uid {
		msgs = c.SelectedMailbox.MessageSetByUID(seqSet)
	} else {
		msgs = c.SelectedMailbox.MessageSetBySequenceNumber(seqSet)
	}

	flagField := types.FlagsFromString(flags)
	for _, msg := range msgs {

		if operation == "+" {
			msg = msg.AddFlags(flagField)
		} else if operation == "-" {
			msg = msg.RemoveFlags(flagField)
		} else {
			msg = msg.OverwriteFlags(flagField)
		}
		msg.Save()

		if err != nil {
			c.writeResponse(args.ID(), "NO "+err.Error())
			return
		}

		// Auto-fetch for the client
		if !silent {
			newFlags, err := fetch("FLAGS", c, msg)
			if err != nil {
				c.writeResponse(args.ID(), "NO "+err.Error())
				return
			}

			fetchResponse := fmt.Sprintf("%d FETCH (%s)",
				msg.SequenceNumber(),
				newFlags,
			)

			c.writeResponse("", fetchResponse)
		}
	}

	c.writeResponse(args.ID(), "OK STORE Completed")
}
