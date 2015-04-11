package imap_server

import (
	"io/ioutil"
	"net/textproto"
	"os"
	"regexp"
	"testing"
)

type rig struct {
	cConn     *textproto.Conn
	sConn     *Conn
	server    *Server
	mailstore DummyMailstore
	inbox     *DummyMailbox
}

func (r *rig) expect(t *testing.T, expected string) {
	line, err := r.cConn.ReadLine()
	if err != nil {
		t.Fatalf("Error reading line: %s", err)
		return
	}
	if line != expected {
		t.Fatalf("Response did not match:\nExpected: %s\nActual:   %s", expected, line)
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
		t.Fatalf("Response did not match pattern:\nExpected: %s\nActual:   %s", pattern, line)
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
		mailstore: server.mailstore.(DummyMailstore),
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

func TestListDirectorySeparator(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 LIST \"\" \"\"")
	r.expect(t, "* LIST (\\Noselect) \"/\" \"\"")
	r.expect(t, "abcd.123 OK LIST completed")
}

func TestListAllMailboxes(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	go r.sConn.Start()
	r.sConn.user = r.mailstore.user
	r.cConn.PrintfLine("abcd.123 LIST \"\" \"*\"")
	r.expect(t, "* LIST () \"/\" \"INBOX\"")
	r.expect(t, "* LIST () \"/\" \"Trash\"")
	r.expect(t, "abcd.123 OK LIST completed")
}

func TestCapabilities(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateNotAuthenticated)
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 CAPABILITY")
	r.expect(t, "* CAPABILITY IMAP4rev1 AUTH=PLAIN")
	r.expect(t, "abcd.123 OK CAPABILITY completed")
}

func TestSelect(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 SELECT INBOX")
	r.expect(t, "* 1 EXISTS")
	r.expect(t, "* 1 RECENT")
	r.expect(t, "* OK [UNSEEN 1]")
	r.expect(t, "* OK [UIDNEXT 2]")
	r.expect(t, "* OK [UIDVALIDITY 250]")
	r.expect(t, "* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)")
}

func TestStatus(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 STATUS INBOX (UIDNEXT UNSEEN)")
	r.expect(t, "* STATUS INBOX (UIDNEXT 2 UNSEEN 1)")
	r.expect(t, "abcd.123 OK STATUS Completed")
}

func TestFetchFlagsUid(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 FETCH 1 (FLAGS UID)")
	r.expect(t, "* 1 FETCH (FLAGS (\\Recent) UID 1)")
	r.expect(t, "abcd.123 OK FETCH Completed")

	// Command case insensitivity
	r.cConn.PrintfLine("abcd.124 fetch 1 (FLAGS)")
	r.expect(t, "* 1 FETCH (FLAGS (\\Recent))")
	r.expect(t, "abcd.124 OK FETCH Completed")
}

func TestFetchHeader(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 FETCH 1 (BODY[HEADER])")
	r.expect(t, "* 1 FETCH (BODY[HEADER] {143}")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expect(t, "")
	r.expect(t, ")")
	r.expect(t, "abcd.123 OK FETCH Completed")
}

func TestFetchSpecificHeaders(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 FETCH 1 (BODY[HEADER.FIELDS (From Subject)])")
	r.expect(t, "* 1 FETCH (BODY[HEADER.FIELDS (\"From\" \"Subject\")] {55}")
	r.expectPattern(t, "^((?i)(subject)|(from)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(from)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expect(t, "")
	r.expect(t, ")")
	r.expect(t, "abcd.123 OK FETCH Completed")
}

func TestFetchPeekSpecificHeaders(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 FETCH 1 (BODY.PEEK[HEADER.FIELDS (from Subject x-priority)])")
	r.expect(t, "* 1 FETCH (BODY[HEADER.FIELDS (\"from\" \"Subject\" \"x-priority\")] {55}")
	r.expectPattern(t, "^((?i)(subject)|(from)): [A-z0-9\\s@\\.]+$")
	r.expectPattern(t, "^((?i)(subject)|(from)): [A-z0-9\\s@\\.]+$")
	r.expect(t, "")
	r.expect(t, ")")
	r.expect(t, "abcd.123 OK FETCH Completed")
}

func TestFetchInternalDate(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 FETCH 1 (INTERNALDATE)")
	r.expect(t, "* 1 FETCH (INTERNALDATE \"28-Oct-2014 00:09:00 +0700\")")
	r.expect(t, "abcd.123 OK FETCH Completed")
}

func TestFetchRFC822Size(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 FETCH 1 (RFC822.SIZE)")
	r.expect(t, "* 1 FETCH (RFC822.SIZE 189)")
	r.expect(t, "abcd.123 OK FETCH Completed")
}

func TestFetchBodyOnly(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 FETCH 1 (BODY[TEXT])")
	r.expect(t, "* 1 FETCH (BODY[TEXT] {54}")
	r.expect(t, "This is the body of the email.")
	r.expect(t, "It is a short email")
	r.expect(t, ")")
	r.expect(t, "abcd.123 OK FETCH Completed")
}

func TestFetchFullMessage(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 FETCH 1 (BODY[])")
	r.expect(t, "* 1 FETCH (BODY[] {195}")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expect(t, "")
	r.expect(t, "This is the body of the email.")
	r.expect(t, "It is a short email")
	r.expect(t, ")")
	r.expect(t, "abcd.123 OK FETCH Completed")
}

func TestFetchFullMessageByUid(t *testing.T) {
	r := setup(t)
	defer r.cleanup()
	r.sConn.setState(StateAuthenticated)
	r.sConn.user = r.mailstore.user
	r.sConn.selectedMailbox = r.mailstore.user.mailboxes[0]
	go r.sConn.Start()
	r.cConn.PrintfLine("abcd.123 UID FETCH 1 (BODY[])")
	r.expect(t, "* 1 FETCH (BODY[] {195}")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expectPattern(t, "^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
	r.expect(t, "")
	r.expect(t, "This is the body of the email.")
	r.expect(t, "It is a short email")
	r.expect(t, " UID 1)")
	r.expect(t, "abcd.123 OK UID FETCH Completed")
}
