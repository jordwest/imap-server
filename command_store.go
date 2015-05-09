package imap

import (
	"fmt"
	"strings"
)

const storeArgUID int = 0
const storeArgRange int = 1
const storeArgOperation int = 2
const storeArgSilent int = 3
const storeArgFlags int = 4

func cmdStoreFlags(args commandArgs, c *Conn) {
	fmt.Printf("STORE command args: %s\n\n", strings.Join(args, ","))
	args.DebugPrint("STORE command")
	operation := args.Arg(storeArgOperation)
	flags := args.Arg(storeArgFlags)
	uid := args.Arg(storeArgUID) == "UID "

	silent := false
	if args.Arg(storeArgSilent) == ".SILENT" {
		silent = true
	}
	if silent {
		fmt.Printf("Silently ")
	}

	if uid {
		fmt.Printf("(searching by UID) ")
	}

	if operation == "+" {
		fmt.Printf("Add flags %s\n", flags)
	} else if operation == "-" {
		fmt.Printf("Remove flags %s\n", flags)
	} else {
		fmt.Printf("Set flags %s\n", flags)
	}

	c.writeResponse(args.ID(), "OK STORE Completed")
}
