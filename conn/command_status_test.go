package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("STATUS Command", func() {
	Context("When logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should respond with the status of INBOX", func() {
			SendLine("abcd.123 STATUS INBOX (UIDNEXT UNSEEN)")
			ExpectResponse("* STATUS INBOX (UIDNEXT 13 UNSEEN 3)")
			ExpectResponse("abcd.123 OK STATUS Completed")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should give an error", func() {
			SendLine("abcd.123 STATUS INBOX (UIDNEXT UNSEEN)")
			ExpectResponse("abcd.123 BAD not authenticated")
		})
	})
})
