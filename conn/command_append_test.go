package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("APPEND Command", func() {
	Context("When a user is logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should append a message with flags and date", func() {
			SendLine("abcd.123 append \"INBOX\" (\\Seen) \"21-Jun-2015 01:00:25 +0900\" {123}")
			ExpectResponse("+ go ahead, feed me your message")
			SendLine("To: you@testing.com")
			SendLine("From: me@testing.com")
			SendLine("Subject: This is a newly appended email")
			SendLine("")
			SendLine("Hello! This is the body.")
			SendLine("From me")
			SendLine("")
			ExpectResponse("abcd.123 OK APPEND completed")

			// Ensure that the email was indeed appended
			mbox := tConn.User.Mailboxes()[0]
			Expect(mbox.Messages()).To(Equal(uint32(4)))
			Expect(mbox.NextUID()).To(Equal(uint32(14)))

			msg := mbox.MessageByUID(13)
			Expect(msg.Header().Get("From")).To(Equal("me@testing.com"))
			Expect(msg.Header().Get("To")).To(Equal("you@testing.com"))
			Expect(msg.Header().Get("Subject")).To(Equal("This is a newly appended email"))

			// Ensure no other emails were interfered with
			msg = mbox.MessageBySequenceNumber(1)
			Expect(msg.Header().Get("Subject")).To(Equal("Test email"))
			msg = mbox.MessageBySequenceNumber(2)
			Expect(msg.Header().Get("Subject")).To(Equal("Another test email"))
			msg = mbox.MessageBySequenceNumber(3)
			Expect(msg.Header().Get("Subject")).To(Equal("Last email"))
			msg = mbox.MessageBySequenceNumber(4)
			Expect(msg.Header().Get("Subject")).To(Equal("This is a newly appended email"))
		})
	})
})
