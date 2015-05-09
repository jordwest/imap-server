package imap

import (
	"errors"
	"fmt"
	"time"
)

// DummyMailstore is an in-memory mail storage for testing purposes and to
// provide an example implementation of a mailstore
type DummyMailstore struct {
	user DummyUser
}

func newDummyMailbox(name string) DummyMailbox {
	return DummyMailbox{
		name:     name,
		messages: make([]Message, 0),
		nextuid:  1,
	}
}

// NewDummyMailstore performs some initialisation and should always be
// used to create a new DummyMailstore
func NewDummyMailstore() DummyMailstore {
	ms := DummyMailstore{
		user: DummyUser{
			authenticated: false,
			mailboxes:     make([]DummyMailbox, 2),
		},
	}
	ms.user.mailboxes[0] = newDummyMailbox("INBOX")
	ms.user.mailboxes[0].addEmail("me@test.com", "you@test.com", "Test email", time.Now(),
		"Test email\r\n"+
			"Regards,\r\n"+
			"Me")
	ms.user.mailboxes[0].addEmail("me@test.com", "you@test.com", "Another test email", time.Now(),
		"Another test email")
	ms.user.mailboxes[0].addEmail("me@test.com", "you@test.com", "Last email", time.Now(),
		"Hello")

	ms.user.mailboxes[1] = newDummyMailbox("Trash")
	return ms
}

// Authenticate implements the Authenticate method on the Mailstore interface
func (d DummyMailstore) Authenticate(username string, password string) (User, error) {
	if username != "username" {
		return DummyUser{}, errors.New("Invalid username. Use 'username'")
	}

	if password != "password" {
		return DummyUser{}, errors.New("Invalid password. Use 'password'")
	}

	d.user.authenticated = true
	return d.user, nil
}

// DummyUser is an in-memory representation of a mailstore's user
type DummyUser struct {
	authenticated bool
	mailboxes     []DummyMailbox
}

// Mailboxes implements the Mailboxes method on the User interface
func (u DummyUser) Mailboxes() []Mailbox {
	mailboxes := make([]Mailbox, len(u.mailboxes))
	index := 0
	for _, element := range u.mailboxes {
		mailboxes[index] = element
		index++
	}
	return mailboxes
}

// MailboxByName returns a DummyMailbox object, given the mailbox's name
func (u DummyUser) MailboxByName(name string) (Mailbox, error) {
	for _, mailbox := range u.mailboxes {
		if mailbox.Name() == name {
			return mailbox, nil
		}
	}
	return DummyMailbox{}, errors.New("Invalid mailbox")
}

// DummyMailbox is an in-memory implementation of a Mailstore Mailbox
type DummyMailbox struct {
	name     string
	nextuid  uint32
	messages []Message
}

// Name returns the Mailbox's name
func (m DummyMailbox) Name() string { return m.name }

// NextUID returns the UID that is likely to be assigned to the next
// new message in the Mailbox
func (m DummyMailbox) NextUID() uint32 { return m.nextuid }

// Recent returns the number of messages in the mailbox which are currently
// marked with the 'Recent' flag
func (m DummyMailbox) Recent() uint32 {
	var count uint32
	for _, message := range m.messages {
		if message.IsRecent() {
			count++
		}
	}
	return count
}

// Messages returns the total number of messages in the Mailbox
func (m DummyMailbox) Messages() uint32 { return uint32(len(m.messages)) }

// Unseen returns the number of messages in the mailbox which are currently
// marked with the 'Unseen' flag
func (m DummyMailbox) Unseen() uint32 {
	count := uint32(0)
	for _, message := range m.messages {
		if !message.IsSeen() {
			count++
		}
	}
	return count
}

// MessageBySequenceNumber returns a single message given the message's sequence number
func (m DummyMailbox) MessageBySequenceNumber(seqno uint32) Message {
	if seqno >= uint32(len(m.messages)) {
		return DummyMessage{}
	}
	return m.messages[seqno-1]
}

// MessageByUID returns a single message given the message's sequence number
func (m DummyMailbox) MessageByUID(uidno uint32) Message {
	for _, message := range m.messages {
		if message.UID() == uidno {
			return message
		}
	}

	// No message found
	return DummyMessage{}
}

// MessageSetByUID returns a slice of messages given a set of UID ranges.
// eg 1,5,9,28:140,190:*
func (m DummyMailbox) MessageSetByUID(set SequenceSet) []Message {
	msgs := make([]Message, 2)
	msgs[0] = m.MessageByUID(1)
	msgs[1] = m.MessageByUID(2)
	return msgs

}

// MessageSetBySequenceNumber returns a slice of messages given a set of
// sequence number ranges
func (m DummyMailbox) MessageSetBySequenceNumber(set SequenceSet) []Message {
	var msgs []Message

	// For each sequence range in the sequence set
	for _, msgRange := range set {
		start, err := msgRange.min.Value()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return msgs
		}

		// If no max is specified, the sequence number must be either a fixed
		// sequence number or
		if msgRange.max.Nil() {
			var sequenceNo uint32
			if msgRange.min.Last() {
				// Fetch the last message in the mailbox
				// (sequence number = total number of messages in mailbox)
				sequenceNo = m.Messages()
			} else {
				// Fetch specific message by sequence number
				sequenceNo, err = msgRange.min.Value()
				if err != nil {
					fmt.Printf("Error: %s\n", err.Error())
					return msgs
				}
			}
			msgs = append(msgs, m.MessageBySequenceNumber(sequenceNo))
			continue
		}

		var end uint32
		if msgRange.max.Last() {
			end = uint32(len(m.messages))
		} else {
			end, err = msgRange.max.Value()
		}

		// Note this is very inefficient when
		// the message array is large. A proper
		// storage system using eg SQL might
		// instead perform a query here using
		// the range values instead.
		for index := uint32(start); index <= end; index++ {
			msgs = append(msgs, m.MessageBySequenceNumber(index-1))
		}
	}
	return msgs

}

func (m *DummyMailbox) addEmail(from string, to string, subject string, date time.Time, body string) {
	uid := m.nextuid
	m.nextuid++

	hdr := make(map[string]string)
	hdr["Date"] = date.Format(rfc822Date)
	hdr["To"] = to
	hdr["From"] = from
	hdr["Subject"] = subject
	hdr["Message-ID"] = fmt.Sprintf("<%d@test.com>", uid)

	newMessage := DummyMessage{
		sequenceNumber: uint32(len(m.messages) + 1),
		uid:            uid,
		recent:         true,
		header:         hdr,
	}
	m.messages = append(m.messages, newMessage)
}

// DummyMessage is a representation of a single in-memory message in a DummyMailbox
type DummyMessage struct {
	sequenceNumber uint32
	uid            uint32
	header         MIMEHeader
	seen           bool
	deleted        bool
	recent         bool
	answered       bool
	flagged        bool
	draft          bool
}

// Header returns the message's MIME Header
func (m DummyMessage) Header() (hdr MIMEHeader) {
	return m.header
}

// UID returns the message's unique identifier (UID)
func (m DummyMessage) UID() uint32 { return m.uid }

// SequenceNumber returns the message's sequence number
func (m DummyMessage) SequenceNumber() uint32 { return m.sequenceNumber }

// Size returns the message's full RFC822 size, including full message header and body
func (m DummyMessage) Size() uint32 {
	hdrStr := fmt.Sprintf("%s\r\n", m.Header())
	return uint32(len(hdrStr)) + uint32(len(m.Body()))
}

// InternalDate returns the internally stored date of the message
func (m DummyMessage) InternalDate() time.Time {
	tz := time.FixedZone("Australia/Brisbane", 10*60*60)
	return time.Date(2014, 10, 28, 0, 9, 0, 0, tz)
}

// Body returns the full body of the message
func (m DummyMessage) Body() string {
	return `This is the body of the email.
It is a short email`
}

// Keywords returns any keywords associated with the message
func (m DummyMessage) Keywords() []string {
	var f []string
	//f[0] = "Test"
	return f
}

// IsSeen returns true if the seen flag is set, false otherwise
func (m DummyMessage) IsSeen() bool { return m.seen }

// IsAnswered returns true if the Answered flag is set, false otherwise
func (m DummyMessage) IsAnswered() bool { return m.answered }

// IsFlagged returns true if the Flagged flag is set, false otherwise
func (m DummyMessage) IsFlagged() bool { return m.flagged }

// IsDeleted returns true if the Deleted flag is set, false otherwise
func (m DummyMessage) IsDeleted() bool { return m.deleted }

// IsDraft returns true if the Draft flag is set, false otherwise
func (m DummyMessage) IsDraft() bool { return m.draft }

// IsRecent returns true if the Recent flag is set, false otherwise
func (m DummyMessage) IsRecent() bool { return m.recent }
