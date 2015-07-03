package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
)

var _ = Describe("APPEND Command", func() {
	Context("When a user is logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should append a message with flags and date", func() {
			SendLine("abcd.123 append \"INBOX\" (\\Seen) \"21-Jun-2015 01:00:25 +0900\" {114}")
			ExpectResponse("+ go ahead, feed me your message")
			//19+2 = 21
			SendLine("To: you@testing.com")
			//20+2 = 22
			SendLine("From: me@testing.com")
			//39+2 = 41
			SendLine("Subject: This is a newly appended email")
			//2
			SendLine("")
			//24+2 = 26
			SendLine("Hello! This is the body.")
			//7+2 = 9
			SendLine("From me")
			//2
			SendLine("")
			ExpectResponse("abcd.123 OK APPEND completed")
		})
	})
})
