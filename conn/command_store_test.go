package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("STORE Command", func() {
	Context("When a mailbox is selected", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateSelected)
			tConn.User = mStore.User
			tConn.SelectedMailbox = tConn.User.Mailboxes()[0]
		})

		It("should silently add a flag to a message", func() {
			SendLine("abcd.123 STORE 1 +FLAGS.SILENT (\\Seen)")
			ExpectResponse("abcd.123 OK STORE Completed")
		})

		It("should remove a flag from a message by UID", func() {
			SendLine("abcd.124 UID STORE 12 -FLAGS (\\Seen)")
			ExpectResponse("* 3 FETCH (FLAGS (\\Recent))")
			ExpectResponse("abcd.124 OK STORE Completed")
		})

		It("should overwrite multiple flags on multiple message by UID", func() {
			SendLine("abcd.125 uid STORE 3:* FLAGS (\\Deleted \\Seen)")
			ExpectResponse("* 1 FETCH (FLAGS (\\Seen \\Deleted))")
			ExpectResponse("* 2 FETCH (FLAGS (\\Seen \\Deleted))")
			ExpectResponse("* 3 FETCH (FLAGS (\\Seen \\Deleted))")
			ExpectResponse("abcd.125 OK STORE Completed")
		})
	})

	Context("When logged in but no mailbox is selected", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should return an error", func() {
			SendLine("abcd.123 STORE 1 FLAGS (\\Seen)")
			ExpectResponse("abcd.123 BAD not selected")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should return an error", func() {
			SendLine("abcd.123 STORE 1 +FLAGS.SILENT (\\Seen)")
			ExpectResponse("abcd.123 BAD not authenticated")
		})
	})
})
