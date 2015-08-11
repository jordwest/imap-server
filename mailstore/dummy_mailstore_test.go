package mailstore

import (
	"testing"

	"github.com/jordwest/imap-server/types"
)

func getDefaultInbox(t *testing.T) *DummyMailbox {
	m := NewDummyMailstore()
	user, err := m.Authenticate("username", "password")
	if err != nil {
		t.Fatalf("Error getting user: %s\n", err)
	}
	mailbox, err := user.MailboxByName("INBOX")
	if err != nil {
		t.Fatalf("Error getting default mailbox: %s\n", err)
	}
	return mailbox.(*DummyMailbox)
}

func assertMessageUIDs(t *testing.T, msgs []Message, uids []uint32) {
	if len(msgs) != len(uids) {
		t.Errorf("Expecting %d messages, got %d messages\n", len(uids), len(msgs))
		debugPrintMessages(msgs)
		return
	}

	errorOccurred := false
	for index, expected := range uids {
		actual := msgs[index].UID()
		if actual != expected {
			t.Errorf("Expected msgs[%d].UID() == %d, got %d\n", index, expected, actual)
			errorOccurred = true
		}
	}

	if errorOccurred {
		debugPrintMessages(msgs)
	}
}

func TestMessageSetBySequenceNumber(t *testing.T) {
	inbox := getDefaultInbox(t)
	msgs := inbox.MessageSetBySequenceNumber(types.SequenceSet{
		types.SequenceRange{Min: "1", Max: ""},
		types.SequenceRange{Min: "4", Max: "*"},
	})
	assertMessageUIDs(t, msgs, []uint32{10})

	msgs = inbox.MessageSetBySequenceNumber(types.SequenceSet{
		types.SequenceRange{Min: "2", Max: "3"},
	})
	assertMessageUIDs(t, msgs, []uint32{11, 12})
}

func TestMessageSetByUID(t *testing.T) {
	inbox := getDefaultInbox(t)
	msgs := inbox.MessageSetByUID(types.SequenceSet{
		types.SequenceRange{Min: "10", Max: "*"},
	})
	assertMessageUIDs(t, msgs, []uint32{10, 11, 12})

	msgs = inbox.MessageSetByUID(types.SequenceSet{
		types.SequenceRange{Min: "3", Max: "9"},
	})
	assertMessageUIDs(t, msgs, []uint32{})

	msgs = inbox.MessageSetByUID(types.SequenceSet{
		types.SequenceRange{Min: "11", Max: "12"},
	})
	assertMessageUIDs(t, msgs, []uint32{11, 12})

	msgs = inbox.MessageSetByUID(types.SequenceSet{
		types.SequenceRange{Min: "*", Max: ""},
	})
	assertMessageUIDs(t, msgs, []uint32{12})
}
