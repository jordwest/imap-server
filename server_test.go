package imap

import (
	"io/ioutil"
	"net/textproto"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/jordwest/imap-server/conn"
	"github.com/jordwest/imap-server/mailstore"
)

type rig struct {
	cConn     *textproto.Conn
	sConn     *conn.Conn
	server    *Server
	mailstore mailstore.DummyMailstore
	inbox     *mailstore.DummyMailbox
}

func (r *rig) expect(t *testing.T, expected string) {
	line, err := r.cConn.ReadLine()
	if err != nil {
		t.Fatalf("Error reading line: %s", err)
		return
	}
	if line != expected {
		t.Fatalf("Response did not match:\nExpected: %s\nActual:	 %s", expected, line)
		return
	}
}

func (r *rig) expectPattern(t *testing.T, pattern string) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("Could not compile pattern: %s\n", pattern)
		return
	}

	line, err := r.cConn.ReadLine()
	if err != nil {
		t.Fatalf("Error reading line: %s", err)
		return
	}
	if re.MatchString(line) == false {
		t.Fatalf("Response did not match pattern:\nExpected: %s\nActual:	 %s", pattern, line)
		return
	}
}

func setup(t *testing.T) rig {
	transcript := ioutil.Discard
	if testing.Verbose() {
		transcript = os.Stdout
	}
	_, cConn, sConn, server, err := NewTestConnection(transcript)
	if err != nil {
		t.Errorf("Error creating test connection: %s", err)
		return rig{}
	}
	return rig{
		sConn:     sConn,
		cConn:     cConn,
		server:    server,
		mailstore: server.mailstore.(mailstore.DummyMailstore),
	}
}

func (r *rig) cleanup() {
	r.cConn.Close()
	r.sConn.Close()
	r.server.Close()
}

func TestWelcomeMessage(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	go r.sConn.Start()
	//cConn.PrintfLine("%s", cmd)
	r.expect(t, "* OK IMAP4rev1 Service Ready")
}

func TestCapabilities(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.SetState(conn.StateNotAuthenticated)
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 CAPABILITY")
	r.expect(t, "* CAPABILITY IMAP4rev1 AUTH=PLAIN")
	r.expect(t, "abcd.123 OK CAPABILITY completed")
}

func TestSelect(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.SetState(conn.StateAuthenticated)
	r.sConn.User = r.mailstore.User
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 SELECT INBOX")
	r.expect(t, "* 3 EXISTS")
	r.expect(t, "* 3 RECENT")
	r.expect(t, "* OK [UNSEEN 3]")
	r.expect(t, "* OK [UIDNEXT 13]")
	r.expect(t, "* OK [UIDVALIDITY 250]")
	r.expect(t, "* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)")
}

func TestStatus(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.SetState(conn.StateAuthenticated)
	r.sConn.User = r.mailstore.User
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 STATUS INBOX (UIDNEXT UNSEEN)")
	r.expect(t, "* STATUS INBOX (UIDNEXT 13 UNSEEN 3)")
	r.expect(t, "abcd.123 OK STATUS Completed")
}

func TestDataRace(t *testing.T) {
	s := NewServer(mailstore.NewDummyMailstore())
	s.Addr = "127.0.0.1:10143"
	go func() {
		s.ListenAndServe()
	}()
	time.Sleep(time.Millisecond)
	s.Close()
}
