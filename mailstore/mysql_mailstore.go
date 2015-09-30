package mailstore

import (
	"errors"
	"fmt"
	"net/textproto"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jordwest/imap-server/types"
)

// HasFlagsSQL returns a MySQL conditional for returning only results
// which have a specified set of flags using bitwise operations
func sqlHasFlags(flag types.Flags) string {
	str := fmt.Sprintf("flags & %d = %d", flag, flag)
	return str
}

func (m *MySQLMailbox) countMessagesWithFlags(flags types.Flags) uint32 {
	count := uint32(0)
	err := m.mailstore.Db.Get(&count, "SELECT COUNT(*) FROM mail_messages WHERE mailbox_id=? AND %s",
		m.ID,
		sqlHasFlags(flags),
	)
	if err != nil {
		panic(err)
	}
	return count
}

// MySQLMailstore is an in-memory mail storage for testing purposes and to
// provide an example implementation of a mailstore
type MySQLMailstore struct {
	User *MySQLUser
	Db   *sqlx.DB
}

// NewMySQLMailstore performs some initialisation and should always be
// used to create a new MySQLMailstore
func NewMySQLMailstore(connectionString string) (*MySQLMailstore, error) {
	db, err := sqlx.Connect("mysql", connectionString)
	ms := &MySQLMailstore{
		Db: db,
		User: &MySQLUser{
			mailstore:     ms,
			authenticated: false,
		},
	}
	return ms, err
}

// Authenticate implements the Authenticate method on the Mailstore interface
func (d *MySQLMailstore) Authenticate(username string, password string) (User, error) {
	if username != "username" {
		return &MySQLUser{}, errors.New("Invalid username. Use 'username'")
	}

	if password != "password" {
		return &MySQLUser{}, errors.New("Invalid password. Use 'password'")
	}

	d.User.authenticated = true
	return d.User, nil
}

// MySQLUser is an in-memory representation of a mailstore's user
type MySQLUser struct {
	authenticated bool
	mailstore     *MySQLMailstore
}

// Mailboxes implements the Mailboxes method on the User interface
func (u *MySQLUser) Mailboxes() []Mailbox {
	result := []MySQLMailbox{}
	err := u.mailstore.Db.Select(&result, "SELECT id, name FROM mail_mailboxes")
	if err != nil {
		panic(err)
	}

	mailboxes := make([]Mailbox, len(result))
	for i, element := range result {
		mailboxes[i] = element
	}
	return mailboxes
}

// MailboxByName returns a MySQLMailbox object, given the mailbox's name
func (u *MySQLUser) MailboxByName(name string) (Mailbox, error) {
	for _, mailbox := range u.mailboxes {
		if mailbox.Name() == name {
			return mailbox, nil
		}
	}
	return nil, errors.New("Invalid mailbox")
}

// MySQLMailbox is an in-memory implementation of a Mailstore Mailbox
type MySQLMailbox struct {
	ID        uint64 `db:"id"`
	name      string `db:"name"`
	nextuid   uint32 `db:"next_uid"`
	mailstore *MySQLMailstore
}

// DebugPrintMailbox prints out all messages in the mailbox to the command line
// for debugging purposes
func (m *MySQLMailbox) DebugPrintMailbox() {
	debugPrintMessages(m.messages)
}

// Name returns the Mailbox's name
func (m *MySQLMailbox) Name() string { return m.name }

// NextUID returns the UID that is likely to be assigned to the next
// new message in the Mailbox
func (m *MySQLMailbox) NextUID() uint32 { return m.nextuid }

// LastUID returns the UID of the last message in the mailbox or if the
// mailbox is empty, the next expected UID
func (m *MySQLMailbox) LastUID() uint32 {
	// TODO
	return 0
}

// Recent returns the number of messages in the mailbox which are currently
// marked with the 'Recent' flag
func (m *MySQLMailbox) Recent() uint32 {
	return m.countMessagesWithFlags(types.FlagRecent)
}

// Messages returns the total number of messages in the Mailbox
func (m *MySQLMailbox) Messages() uint32 {
	return m.countMessagesWithFlags(0)
}

// Unseen returns the number of messages in the mailbox which are currently
// marked with the 'Unseen' flag
func (m *MySQLMailbox) Unseen() uint32 {
	return m.countMessagesWithFlags(types.FlagSeen)
}

// MessageBySequenceNumber returns a single message given the message's sequence number
func (m *MySQLMailbox) MessageBySequenceNumber(seqno uint32) Message {
	if seqno > uint32(len(m.messages)) {
		return nil
	}
	return m.messages[seqno-1]
}

// MessageByUID returns a single message given the message's sequence number
func (m *MySQLMailbox) MessageByUID(uidno uint32) Message {
	for _, message := range m.messages {
		if message.UID() == uidno {
			return message
		}
	}

	// No message found
	return nil
}

// MessageSetByUID returns a slice of messages given a set of UID ranges.
// eg 1,5,9,28:140,190:*
func (m *MySQLMailbox) MessageSetByUID(set types.SequenceSet) []Message {
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
			msg := m.MessageByUID(uid)
			if err != nil {
				fmt.Printf("Error: %s\n", err.Error())
				return msgs
			}
			if msg != nil {
				msgs = append(msgs, msg)
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
func (m *MySQLMailbox) MessageSetBySequenceNumber(set types.SequenceSet) []Message {
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
			msg := m.MessageBySequenceNumber(sequenceNo)
			if msg != nil {
				msgs = append(msgs, msg)
			}
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

// NewMessage creates a new message in the dummy mailbox.
func (m *MySQLMailbox) NewMessage() Message {
	return &MySQLMessage{
		sequenceNumber: 0,
		uid:            0,
		header:         make(textproto.MIMEHeader),
		internalDate:   time.Now(),
		flags:          types.Flags(0),
		mailstore:      m.mailstore,
		mailboxID:      m.ID,
		body:           "",
	}
}

// MySQLMessage is a representation of a single in-memory message in a
// MySQLMailbox.
type MySQLMessage struct {
	sequenceNumber uint32               `db:"sequence_number"`
	uid            uint32               `db:"uid"`
	header         textproto.MIMEHeader `db:"header"`
	internalDate   time.Time            `db:"date"`
	flags          types.Flags          `db:"flags"`
	mailboxID      uint32               `db:"mailbox_id"`
	body           string               `db:"body"`
	mailstore      *MySQLMailstore
}

// Header returns the message's MIME Header.
func (m *MySQLMessage) Header() (hdr textproto.MIMEHeader) {
	return m.header
}

// UID returns the message's unique identifier (UID).
func (m *MySQLMessage) UID() uint32 { return m.uid }

// SequenceNumber returns the message's sequence number.
func (m *MySQLMessage) SequenceNumber() uint32 { return m.sequenceNumber }

// Size returns the message's full RFC822 size, including full message header
// and body.
func (m *MySQLMessage) Size() uint32 {
	hdrStr := fmt.Sprintf("%s\r\n", m.Header())
	return uint32(len(hdrStr)) + uint32(len(m.Body()))
}

// InternalDate returns the internally stored date of the message
func (m *MySQLMessage) InternalDate() time.Time {
	return m.internalDate
}

// Body returns the full body of the message
func (m *MySQLMessage) Body() string {
	return m.body
}

// Keywords returns any keywords associated with the message
func (m *MySQLMessage) Keywords() []string {
	var f []string
	//f[0] = "Test"
	return f
}

// Flags returns any flags on the message.
func (m *MySQLMessage) Flags() types.Flags {
	return m.flags
}

// OverwriteFlags replaces any flags on the message with those specified.
func (m *MySQLMessage) OverwriteFlags(newFlags types.Flags) Message {
	m.flags = newFlags
	return m
}

// AddFlags adds the given flag to the message.
func (m *MySQLMessage) AddFlags(newFlags types.Flags) Message {
	m.flags = m.flags.SetFlags(newFlags)
	return m
}

// RemoveFlags removes the given flag from the message.
func (m *MySQLMessage) RemoveFlags(newFlags types.Flags) Message {
	m.flags = m.flags.ResetFlags(newFlags)
	return m
}

// SetHeaders sets the e-mail headers of the message.
func (m *MySQLMessage) SetHeaders(newHeader textproto.MIMEHeader) Message {
	m.header = newHeader
	return m
}

// SetBody sets the body of the message.
func (m *MySQLMessage) SetBody(newBody string) Message {
	m.body = newBody
	return m
}

// Save saves the message to the mailbox it belongs to.
func (m *MySQLMessage) Save() (Message, error) {
	mailbox := m.mailstore.User.mailboxes[m.mailboxID]
	if m.sequenceNumber == 0 {
		// Message is new
		m.uid = mailbox.nextuid
		mailbox.nextuid++
		m.sequenceNumber = uint32(len(mailbox.messages))
		mailbox.messages = append(mailbox.messages, m)
	} else {
		// Message exists
		mailbox.messages[m.sequenceNumber-1] = m
	}
	return m, nil
}

// DeleteFlaggedMessages deletes messages marked with the Delete flag and
// returns them.
func (m *MySQLMailbox) DeleteFlaggedMessages() ([]Message, error) {
	var delIDs []int
	var delMsgs []Message

	// Find messages to be deleted.
	for i, msg := range m.messages {
		if msg.Flags().HasFlags(types.FlagDeleted) {
			delIDs = append(delIDs, i)
			delMsgs = append(delMsgs, msg)
		}
	}

	// Delete message from slice. Run this backward because otherwise it would
	// fail if we have multiple items to remove.
	for x := len(delIDs) - 1; x >= 0; x-- {
		i := delIDs[x]
		// From: https://github.com/golang/go/wiki/SliceTricks
		m.messages, m.messages[len(m.messages)-1] = append(m.messages[:i],
			m.messages[i+1:]...), nil
	}

	// Update sequence numbers.
	for i, msg := range m.messages {
		dmsg := msg.(*MySQLMessage)
		dmsg.sequenceNumber = uint32(i) + 1
	}

	return delMsgs, nil
}
