package main

import (
	"fmt"
	"os"

	imap_server "github.com/jordwest/imap-server"
)

func main() {
	s := imap_server.NewServer()
	s.Transcript = os.Stdout
	s.Addr = ":10143"

	err := s.ListenAndServe()
	if err != nil {
		fmt.Printf("Error creating test connection: %s\n", err)
	}
}
