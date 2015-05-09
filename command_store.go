package imap

import "fmt"

const storeArgRange int = 2
const storeArgOperation int = 3
const storeArgSilent int = 4
const storeArgFlags int = 5

func cmdStoreFlags(args commandArgs, c *Conn) {
	fmt.Printf("STORE command args: %+v\n\n", args)
	operation := args.Arg(storeArgOperation)
	flags := args.Arg(storeArgFlags)

	silent := false
	if args.Arg(storeArgSilent) == ".SILENT" {
		silent = true
	}
	if silent {
		fmt.Printf("Silently ")
	}

	if operation == "+" {
		fmt.Printf("Add flags %s\n", flags)
	} else if operation == "-" {
		fmt.Printf("Remove flags %s\n", flags)
	} else {
		fmt.Printf("Set flags %s\n", flags)
	}

	c.writeResponse(args.Id(), "OK STORE Completed")
}
