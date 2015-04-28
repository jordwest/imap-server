package imap_server

import "fmt"

const STORE_ARG_OPERATION int = 3
const STORE_ARG_SILENT int = 4
const STORE_ARG_FLAGS int = 5

func cmdStoreFlags(args commandArgs, c *Conn) {
	fmt.Printf("STORE command args: %+v\n\n", args)
	operation := args.Arg(STORE_ARG_OPERATION)
	flags := args.Arg(STORE_ARG_FLAGS)

	silent := false
	if args.Arg(STORE_ARG_SILENT) == ".SILENT" {
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
