package imap

import (
	"net"
	"testing"
	"time"

	"github.com/jordwest/imap-server/mailstore"
)

func TestDataRace(t *testing.T) {
	s := NewServer(mailstore.NewDummyMailstore())
	addr := "127.0.0.1:10143"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		s.Serve(l)
	}()
	time.Sleep(time.Millisecond)
	l.Close()
}
