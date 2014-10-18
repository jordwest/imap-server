package main

import (
	"fmt"
	"io"
	"os"

	"github.com/jordwest/imap-server"
)

func main() {
	_, _, sConn, err := imap_server.NewTestConnection()
	var stdout io.Writer
	stdout = os.Stdout
	fmt.Printf("Stdout is %+v\n", stdout)
	sConn.Transcript = stdout
	if err != nil {
		fmt.Printf("Error creating test connection: %s\n", err)
	}
	sConn.Start()
}
