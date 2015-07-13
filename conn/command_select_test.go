package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("SELECT Command", func() {
	Context("When logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should select the INBOX mailbox", func() {
			SendLine("abcd.123 SELECT INBOX")
			ExpectResponse("* 3 EXISTS")
			ExpectResponse("* 3 RECENT")
			ExpectResponse("* OK [UNSEEN 3]")
			ExpectResponse("* OK [UIDNEXT 13]")
			ExpectResponse("* OK [UIDVALIDITY 250]")
			ExpectResponse("* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should give an error", func() {
			SendLine("abcd.123 SELECT INBOX")
			ExpectResponse("abcd.123 BAD not authenticated")
		})
	})
})
