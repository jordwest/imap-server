package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("LOGOUT Command", func() {
	Context("When logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should log the user out", func() {
			SendLine("abcd.123 LOGOUT")
			ExpectResponse("* BYE IMAP4rev1 server logging out")
			ExpectResponse("abcd.123 OK LOGOUT completed")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should give an error", func() {
			SendLine("abcd.123 LOGOUT")
			ExpectResponse("* BYE IMAP4rev1 server logging out")
			ExpectResponse("abcd.123 OK LOGOUT completed")
		})
	})
})
