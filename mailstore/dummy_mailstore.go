package mailstore

import (
	"errors"
	"fmt"
	"time"

	"github.com/jordwest/imap-server/types"
	"github.com/jordwest/imap-server/util"
)

// DummyMailstore is an in-memory mail storage for testing purposes and to
// provide an example implementation of a mailstore
type DummyMailstore struct {
	User DummyUser
}

func newDummyMailbox(name string) DummyMailbox {
	return DummyMailbox{
		name:     name,
		messages: make([]Message, 0),
		nextuid:  10,
	}
}

// NewDummyMailstore performs some initialisation and should always be
// used to create a new DummyMailstore
func NewDummyMailstore() DummyMailstore {
	ms := DummyMailstore{
		User: DummyUser{
			authenticated: false,
			mailboxes:     make([]DummyMailbox, 2),
		},
	}
	ms.User.mailboxes[0] = newDummyMailbox("INBOX")
	ms.User.mailboxes[0].addEmail("me@test.com", "you@test.com", "Test email", time.Now(),
		"Test email\r\n"+
			"Regards,\r\n"+
			"Me")
	ms.User.mailboxes[0].addEmail("me@test.com", "you@test.com", "Another test email", time.Now(),
		"Another test email")
	ms.User.mailboxes[0].addEmail("me@test.com", "you@test.com", "Last email", time.Now(),
		"Hello")

	ms.User.mailboxes[1] = newDummyMailbox("Trash")
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

	d.User.authenticated = true
	return d.User, nil
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

// DebugPrintMailbox prints out all messages in the mailbox to the command line
// for debugging purposes
func (m DummyMailbox) DebugPrintMailbox() {
	debugPrintMessages(m.messages)
}

// Name returns the Mailbox's name
func (m DummyMailbox) Name() string { return m.name }

// NextUID returns the UID that is likely to be assigned to the next
// new message in the Mailbox
func (m DummyMailbox) NextUID() uint32 { return m.nextuid }

// LastUID returns the UID of the last message in the mailbox or if the
// mailbox is empty, the next expected UID
func (m DummyMailbox) LastUID() uint32 {
	lastMsgIndex := len(m.messages) - 1

	// If no messages in the mailbox, return the next UID
	if lastMsgIndex == -1 {
		return m.NextUID()
	}

	return m.messages[lastMsgIndex].UID()
}

// Recent returns the number of messages in the mailbox which are currently
// marked with the 'Recent' flag
func (m DummyMailbox) Recent() uint32 {
	var count uint32
	for _, message := range m.messages {
		if message.Flags().HasFlags(types.FlagRecent) {
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
		if !message.Flags().HasFlags(types.FlagSeen) {
			count++
		}
	}
	return count
}

// MessageBySequenceNumber returns a single message given the message's sequence number
func (m DummyMailbox) MessageBySequenceNumber(seqno uint32) Message {
	if seqno > uint32(len(m.messages)) {
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
func (m DummyMailbox) MessageSetByUID(set types.SequenceSet) []Message {
	var msgs []Message

	// If the mailbox is empty, return empty array
	if m.Messages() == 0 {
		return msgs
	}

	for _, msgRange := range set {
		// If Min is "*", meaning the last UID in the mailbox, Max should
		// always be Nil
		if msgRange.Min.Last() {
			// Return the last message in the mailbox
			msgs = append(msgs, m.MessageByUID(m.LastUID()))
			continue
		}

		start, err := msgRange.Min.Value()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return msgs
		}

		// If no Max is specified, the sequence number must be either a fixed
		// sequence number or
		if msgRange.Max.Nil() {
			var uid uint32
			// Fetch specific message by sequence number
			uid, err = msgRange.Min.Value()
			msgs = append(msgs, m.MessageByUID(uid))
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return msgs
			}
			continue
		}

		var end uint32
		if msgRange.Max.Last() {
			end = m.LastUID()
		} else {
			end, err = msgRange.Max.Value()
		}

		// Note this is very inefficient when
		// the message array is large. A proper
		// storage system using eg SQL might
		// instead perform a query here using
		// the range values instead.
		for _, msg := range m.messages {
			uid := msg.UID()
			if uid >= start && uid <= end {
				msgs = append(msgs, msg)
			}
		}
		for index := uint32(start); index <= end; index++ {
		}
	}

	return msgs
}

// MessageSetBySequenceNumber returns a slice of messages given a set of
// sequence number ranges
func (m DummyMailbox) MessageSetBySequenceNumber(set types.SequenceSet) []Message {
	var msgs []Message

	// If the mailbox is empty, return empty array
	if m.Messages() == 0 {
		return msgs
	}

	// For each sequence range in the sequence set
	for _, msgRange := range set {
		// If Min is "*", meaning the last message in the mailbox, Max should
		// always be Nil
		if msgRange.Min.Last() {
			// Return the last message in the mailbox
			msgs = append(msgs, m.MessageBySequenceNumber(m.Messages()))
			continue
		}

		start, err := msgRange.Min.Value()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return msgs
		}

		// If no Max is specified, the sequence number must be either a fixed
		// sequence number or
		if msgRange.Max.Nil() {
			var sequenceNo uint32
			// Fetch specific message by sequence number
			sequenceNo, err = msgRange.Min.Value()
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return msgs
			}
			msgs = append(msgs, m.MessageBySequenceNumber(sequenceNo))
			continue
		}

		var end uint32
		if msgRange.Max.Last() {
			end = uint32(len(m.messages))
		} else {
			end, err = msgRange.Max.Value()
		}

		// Note this is very inefficient when
		// the message array is large. A proper
		// storage system using eg SQL might
		// instead perform a query here using
		// the range values instead.
		for seqNo := start; seqNo <= end; seqNo++ {
			msgs = append(msgs, m.MessageBySequenceNumber(seqNo))
		}
	}
	return msgs

}

func (m *DummyMailbox) addEmail(from string, to string, subject string, date time.Time, body string) {
	uid := m.nextuid
	m.nextuid++

	hdr := make(map[string]string)
	hdr["Date"] = date.Format(util.RFC822Date)
	hdr["To"] = to
	hdr["From"] = from
	hdr["Subject"] = subject
	hdr["Message-ID"] = fmt.Sprintf("<%d@test.com>", uid)

	newMessage := DummyMessage{
		sequenceNumber: uint32(len(m.messages) + 1),
		uid:            uid,
		header:         hdr,
	}
	newMessage = newMessage.AddFlags(types.FlagRecent).(DummyMessage)
	newMessage.mailbox = m
	m.messages = append(m.messages, newMessage)
}

// DummyMessage is a representation of a single in-memory message in a DummyMailbox
type DummyMessage struct {
	sequenceNumber uint32
	uid            uint32
	header         types.MIMEHeader
	flags          types.Flags
	mailbox        *DummyMailbox
}

// Header returns the message's MIME Header
func (m DummyMessage) Header() (hdr types.MIMEHeader) {
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

func (m DummyMessage) Flags() types.Flags {
	return m.flags
}

func (m DummyMessage) OverwriteFlags(newFlags types.Flags) Message {
	m.flags = newFlags
	return m
}

func (m DummyMessage) AddFlags(newFlags types.Flags) Message {
	m.flags = m.flags.SetFlags(newFlags)
	return m
}

func (m DummyMessage) RemoveFlags(newFlags types.Flags) Message {
	m.flags = m.flags.ResetFlags(newFlags)
	return m
}

func (m DummyMessage) Save() error {
	m.mailbox.messages[m.sequenceNumber-1] = m
	return nil
}

func debugPrintMessages(messages []Message) {
	fmt.Printf("SeqNo  |UID    |From      |To        |Subject\n")
	fmt.Printf("-------+-------+----------+----------+-------\n")
	for _, msg := range messages {
		_, from, _ := msg.Header().FindKey("from")
		_, to, _ := msg.Header().FindKey("to")
		_, subject, _ := msg.Header().FindKey("subject")
		fmt.Printf("%-7d|%-7d|%-10.10s|%-10.10s|%s\n", msg.SequenceNumber(), msg.UID(), from, to, subject)
	}
}
