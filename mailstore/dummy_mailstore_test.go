package mailstore

import "testing"

func getDefaultInbox(t *testing.T) DummyMailbox {
	m := NewDummyMailstore()
	user, err := m.Authenticate("username", "password")
	if err != nil {
		t.Fatalf("Error getting user: %s\n", err)
	}
	mailbox, err := user.MailboxByName("INBOX")
	if err != nil {
		t.Fatalf("Error getting default mailbox: %s\n", err)
	}
	return mailbox.(DummyMailbox)
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
	msgs := inbox.MessageSetBySequenceNumber(SequenceSet{
		SequenceRange{min: "1", max: ""},
		SequenceRange{min: "4", max: "*"},
	})
	assertMessageUIDs(t, msgs, []uint32{10})

	msgs = inbox.MessageSetBySequenceNumber(SequenceSet{
		SequenceRange{min: "2", max: "3"},
	})
	assertMessageUIDs(t, msgs, []uint32{11, 12})
}

func TestMessageSetByUID(t *testing.T) {
	inbox := getDefaultInbox(t)
	msgs := inbox.MessageSetByUID(SequenceSet{
		SequenceRange{min: "10", max: "*"},
	})
	assertMessageUIDs(t, msgs, []uint32{10, 11, 12})

	msgs = inbox.MessageSetByUID(SequenceSet{
		SequenceRange{min: "3", max: "9"},
	})
	assertMessageUIDs(t, msgs, []uint32{})

	msgs = inbox.MessageSetByUID(SequenceSet{
		SequenceRange{min: "11", max: "12"},
	})
	assertMessageUIDs(t, msgs, []uint32{11, 12})

	msgs = inbox.MessageSetByUID(SequenceSet{
		SequenceRange{min: "*", max: ""},
	})
	assertMessageUIDs(t, msgs, []uint32{12})
}
