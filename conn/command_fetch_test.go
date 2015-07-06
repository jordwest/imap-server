package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"

	//. "github.com/onsi/gomega"
)

var _ = Describe("FETCH Command", func() {
	Context("When a mailbox is selected", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateSelected)
			tConn.SetReadWrite()
			tConn.User = mStore.User
			tConn.SelectedMailbox = tConn.User.Mailboxes()[0]
		})

		It("should fetch the flags from a message by UID", func() {
			SendLine("abcd.123 FETCH 1 (FLAGS UID)")
			ExpectResponse("* 1 FETCH (FLAGS (\\Recent) UID 10)")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should be case insensitive", func() {
			SendLine("abcd.123 fetch 1 (FLAGS UID)")
			ExpectResponse("* 1 FETCH (FLAGS (\\Recent) UID 10)")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should fetch the header of a message", func() {
			SendLine("abcd.123 FETCH 1 (BODY[HEADER])")
			ExpectResponse("* 1 FETCH (BODY[HEADER] {126}")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponse("")
			ExpectResponse(")")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should fetch specific headers of a message", func() {
			SendLine("abcd.123 FETCH 1 (BODY[HEADER.FIELDS (From Subject)])")
			ExpectResponse("* 1 FETCH (BODY[HEADER.FIELDS (\"From\" \"Subject\")] {40}")
			ExpectResponsePattern("^((?i)(subject)|(from)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(from)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponse(")")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should PEEK specific headers of a message without changing the Recent flag", func() {
			SendLine("abcd.123 FETCH 1 (BODY.PEEK[HEADER.FIELDS (from Subject x-priority)])")
			ExpectResponse("* 1 FETCH (BODY[HEADER.FIELDS (\"from\" \"Subject\" \"x-priority\")] {40}")
			ExpectResponsePattern("^((?i)(subject)|(from)): [A-z0-9\\s@\\.]+$")
			ExpectResponsePattern("^((?i)(subject)|(from)): [A-z0-9\\s@\\.]+$")
			ExpectResponse(")")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should fetch the internal date of a message", func() {
			SendLine("abcd.123 FETCH 1 (INTERNALDATE)")
			ExpectResponse("* 1 FETCH (INTERNALDATE \"28-Oct-2014 00:09:00 +0700\")")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should fetch the RFC822 size of a message", func() {
			SendLine("abcd.123 FETCH 1 (RFC822.SIZE)")
			ExpectResponse("* 1 FETCH (RFC822.SIZE 154)")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should fetch the body of a message only", func() {
			SendLine("abcd.123 FETCH 1 (BODY[TEXT])")
			ExpectResponse("* 1 FETCH (BODY[TEXT] {26}")
			ExpectResponse("Test email")
			ExpectResponse("Regards,")
			ExpectResponse("Me")
			ExpectResponse(")")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should fetch a complete message", func() {
			SendLine("abcd.123 FETCH 1 (BODY[])")
			ExpectResponse("* 1 FETCH (BODY[] {152}")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponse("")
			ExpectResponse("Test email")
			ExpectResponse("Regards,")
			ExpectResponse("Me")
			ExpectResponse(")")
			ExpectResponse("abcd.123 OK FETCH Completed")
		})

		It("should fetch a complete message by UID", func() {
			SendLine("abcd.123 UID FETCH 11 (BODY[])")
			ExpectResponse("* 2 FETCH (BODY[] {154}")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponsePattern("^((?i)(subject)|(message-id)|(to)|(from)|(date)): [<>A-z0-9\\s@\\.,\\:\\+]+$")
			ExpectResponse("")
			ExpectResponse("Another test email")
			ExpectResponse(" UID 11)")
			ExpectResponse("abcd.123 OK UID FETCH Completed")
		})

	})

	Context("When logged in but no mailbox is selected", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should return an error", func() {
			SendLine("abcd.123 FETCH 1 (BODY[TEXT])")
			ExpectResponse("abcd.123 BAD not selected")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should return an error", func() {
			SendLine("abcd.123 FETCH 1 (BODY[TEXT])")
			ExpectResponse("abcd.123 BAD not authenticated")
		})
	})
})
