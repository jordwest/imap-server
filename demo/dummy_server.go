package main

import (
	"fmt"
	"os"

	imap "github.com/jordwest/imap-server"
	"github.com/jordwest/imap-server/mailstore"
)

func main() {
	store := mailstore.NewDummyMailstore()
	s := imap.NewServer(store)
	s.Transcript = os.Stdout
	s.Addr = ":10143"

	err := s.ListenAndServe()
	if err != nil {
		fmt.Printf("Error creating test connection: %s\n", err)
	}
}
