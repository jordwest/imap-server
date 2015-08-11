package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	"github.com/jordwest/imap-server/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EXPUNGE Command", func() {
	Context("When a mailbox is selected", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateSelected)
			tConn.SetReadWrite()
			tConn.User = mStore.User
			tConn.SelectedMailbox = tConn.User.Mailboxes()[0]
		})

		It("should delete the second message", func() {
			// Mark the last message as deleted.
			_, err := tConn.SelectedMailbox.MessageBySequenceNumber(2).
				AddFlags(types.FlagDeleted).Save()
			Expect(err).ToNot(HaveOccurred())

			// Expunge the e-mail.
			SendLine("abc.123 EXPUNGE")
			ExpectResponse("* 2 EXPUNGE")
			ExpectResponse("abc.123 OK EXPUNGE completed")

			// Make sure that it was removed from the mailbox.
			Expect(tConn.SelectedMailbox.Messages()).To(Equal(uint32(2)))

			// Make sure that sequence numbers were properly arranged.
			for i := uint32(1); i <= 2; i++ {
				msg := tConn.SelectedMailbox.MessageBySequenceNumber(i)
				Expect(msg.SequenceNumber()).To(Equal(i))
			}
		})

		It("should delete the first and third messages", func() {
			// Mark the first message as deleted.
			_, err := tConn.SelectedMailbox.MessageBySequenceNumber(1).
				AddFlags(types.FlagDeleted).Save()
			Expect(err).ToNot(HaveOccurred())

			// Mark the second message as deleted.
			_, err = tConn.SelectedMailbox.MessageBySequenceNumber(3).
				AddFlags(types.FlagDeleted).Save()
			Expect(err).ToNot(HaveOccurred())

			// Expunge the e-mail.
			SendLine("abc.234 EXPUNGE")
			ExpectResponse("* 1 EXPUNGE")
			ExpectResponse("* 3 EXPUNGE")
			ExpectResponse("abc.234 OK EXPUNGE completed")

			// Make sure that the two messages were deleted from the mailbox.
			Expect(tConn.SelectedMailbox.Messages()).To(Equal(uint32(1)))

			// Make sure that sequence numbers were properly arranged.
			for i := uint32(1); i <= 1; i++ {
				msg := tConn.SelectedMailbox.MessageBySequenceNumber(i)
				Expect(msg.SequenceNumber()).To(Equal(i))
			}
		})

	})

	Context("When logged in but no mailbox is selected", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should return an error", func() {
			SendLine("abcd.123 EXPUNGE")
			ExpectResponse("abcd.123 BAD not selected")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should return an error", func() {
			SendLine("abcd.123 EXPUNGE")
			ExpectResponse("abcd.123 BAD not authenticated")
		})
	})
})
