package imap

import (
	"testing"
	"time"

	"github.com/jordwest/imap-server/mailstore"
)

func TestDataRace(t *testing.T) {
	s := NewServer(mailstore.NewDummyMailstore())
	s.Addr = "127.0.0.1:10143"
	go func() {
		s.ListenAndServe()
	}()
	time.Sleep(time.Millisecond)
	s.Close()
}
