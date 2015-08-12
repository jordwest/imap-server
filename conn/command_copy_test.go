package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	"github.com/jordwest/imap-server/mailstore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("COPY Command", func() {
	Context("When a mailbox is selected", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateSelected)
			tConn.SetReadWrite()
			tConn.User = mStore.User
			tConn.SelectedMailbox = tConn.User.Mailboxes()[0]
		})

		checkMsgEqual := func(actualMsg, expectMsg mailstore.Message) {
			Expect(expectMsg.Body()).To(Equal(actualMsg.Body()))
			Expect(expectMsg.Header()).To(Equal(actualMsg.Header()))
			Expect(expectMsg.Flags()).To(Equal(actualMsg.Flags()))
		}

		It("should copy a message by sequence number", func() {
			SendLine(`abcd.123 COPY 1:2 "Trash"`)
			ExpectResponse("abcd.123 OK COPY Completed")

			// Verify that everything matches.
			trash, _ := tConn.User.MailboxByName("Trash")
			checkMsgEqual(trash.MessageBySequenceNumber(1),
				tConn.SelectedMailbox.MessageBySequenceNumber(1))
			checkMsgEqual(trash.MessageBySequenceNumber(2),
				tConn.SelectedMailbox.MessageBySequenceNumber(2))
		})

		It("should copy a message by UID", func() {
			SendLine(`abcd.123 uid COPY 11:12 "Trash"`)
			ExpectResponse("abcd.123 OK UID COPY Completed")

			// Verify that everything matches.
			trash, _ := tConn.User.MailboxByName("Trash")
			checkMsgEqual(trash.MessageByUID(10),
				tConn.SelectedMailbox.MessageByUID(11))
			checkMsgEqual(trash.MessageByUID(11),
				tConn.SelectedMailbox.MessageByUID(12))
		})

		It("shouldn't copy nonexistant message by sequence number", func() {
			SendLine("abcd.125 uid COPY 50:* Trash")
			ExpectResponse("abcd.125 NO no messages found")
		})

		It("shouldn't copy nonexistant message by UID", func() {
			SendLine("abcd.125 COPY 5:* Trash")
			ExpectResponse("abcd.125 NO no messages found")
		})
	})

	Context("When logged in but no mailbox is selected", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})

		It("should return an error", func() {
			SendLine("abcd.123 COPY 1 INBOX")
			ExpectResponse("abcd.123 BAD not selected")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})

		It("should return an error", func() {
			SendLine("abcd.123 COPY 1 INBOX")
			ExpectResponse("abcd.123 BAD not authenticated")
		})
	})
})
