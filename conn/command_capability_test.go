package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("CAPABILITY Command", func() {
	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should return server capabilities", func() {
			SendLine("abcd.123 CAPABILITY")
			ExpectResponse("* CAPABILITY IMAP4rev1 AUTH=PLAIN")
			ExpectResponse("abcd.123 OK CAPABILITY completed")
		})
	})

})
