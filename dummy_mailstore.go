package imap

import (
	"errors"
	"fmt"
	"time"
)

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

type DummyUser struct {
	authenticated bool
	mailboxes     []DummyMailbox
}

func (u DummyUser) Mailboxes() []Mailbox {
	mailboxes := make([]Mailbox, len(u.mailboxes))
	index := 0
	for _, element := range u.mailboxes {
		mailboxes[index] = element
		index++
	}
	return mailboxes
}

func (u DummyUser) MailboxByName(name string) (Mailbox, error) {
	for _, mailbox := range u.mailboxes {
		if mailbox.Name() == name {
			return mailbox, nil
		}
	}
	return DummyMailbox{}, errors.New("Invalid mailbox")
}

type DummyMailbox struct {
	name     string
	nextuid  uint32
	messages []Message
}

func (m DummyMailbox) Name() string    { return m.name }
func (m DummyMailbox) NextUid() uint32 { return m.nextuid }
func (m DummyMailbox) Recent() uint32 {
	var count uint32 = 0
	for _, message := range m.messages {
		if message.IsRecent() {
			count++
		}
	}
	return count
}
func (m DummyMailbox) Messages() uint32 { return uint32(len(m.messages)) }
func (m DummyMailbox) Unseen() uint32 {
	var count uint32 = 0
	for _, message := range m.messages {
		if !message.IsSeen() {
			count++
		}
	}
	return count
}

func (m DummyMailbox) MessageBySequenceNumber(seqno uint32) Message {
	if seqno >= uint32(len(m.messages)) {
		return DummyMessage{}
	}
	return m.messages[seqno-1]
}

func (m DummyMailbox) MessageByUid(uidno uint32) Message {
	for _, message := range m.messages {
		if message.Uid() == uidno {
			return message
		}
	}

	// No message found
	return DummyMessage{}
}

func (m DummyMailbox) MessageSetByUid(set SequenceSet) []Message {
	msgs := make([]Message, 2)
	msgs[0] = m.MessageByUid(1)
	msgs[1] = m.MessageByUid(2)
	return msgs

}

func (m DummyMailbox) MessageSetBySequenceNumber(set SequenceSet) []Message {
	msgs := make([]Message, 0)

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
		for index := uint32(start); index <= end; index += 1 {
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

func (m DummyMessage) Header() (hdr MIMEHeader) {
	return m.header
}

func (m DummyMessage) Uid() uint32            { return m.uid }
func (m DummyMessage) SequenceNumber() uint32 { return m.sequenceNumber }

func (m DummyMessage) Size() uint32 {
	hdrStr := fmt.Sprintf("%s\r\n", m.Header())
	return uint32(len(hdrStr)) + uint32(len(m.Body()))
}

func (m DummyMessage) InternalDate() time.Time {
	tz := time.FixedZone("Australia/Brisbane", 10*60*60)
	return time.Date(2014, 10, 28, 0, 9, 0, 0, tz)
}

func (m DummyMessage) Body() string {
	return `This is the body of the email.
It is a short email`
}

func (m DummyMessage) Keywords() []string {
	f := make([]string, 0)
	//f[0] = "Test"
	return f
}

func (m DummyMessage) IsSeen() bool     { return m.seen }
func (m DummyMessage) IsAnswered() bool { return m.answered }
func (m DummyMessage) IsFlagged() bool  { return m.flagged }
func (m DummyMessage) IsDeleted() bool  { return m.deleted }
func (m DummyMessage) IsDraft() bool    { return m.draft }
func (m DummyMessage) IsRecent() bool   { return m.recent }
