package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("LIST Command", func() {
	Context("When logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should return the directory separator", func() {
			SendLine("abcd.123 LIST \"\" \"\"")
			ExpectResponse("* LIST (\\Noselect) \"/\" \"\"")
			ExpectResponse("abcd.123 OK LIST completed")
		})

		It("should return the list of mailboxes", func() {
			SendLine("abcd.123 LIST \"\" \"*\"")
			ExpectResponse("* LIST () \"/\" \"INBOX\"")
			ExpectResponse("* LIST () \"/\" \"Trash\"")
			ExpectResponse("abcd.123 OK LIST completed")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should give an error", func() {
			SendLine("abcd.123 LIST \"\" \"*\"")
			ExpectResponse("abcd.123 BAD not authenticated")
		})
	})
})
