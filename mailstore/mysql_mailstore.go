package mailstore

import (
	"errors"
	"fmt"
	"net/textproto"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jordwest/imap-server/types"
)

func (m MySQLMailbox) countMessagesWithFlags(flags types.Flags) uint32 {
	count := uint32(0)
	err := m.mailstore.Db.Get(&count,
		`SELECT COUNT(*)
		 FROM mail_messages
		 WHERE mailbox_id=?
			AND flags & ? = ?`,
		m.ID,
		flags,
		flags,
	)
	if err != nil {
		panic(err)
	}
	return count
}

func (m MySQLMailbox) countMessagesWithoutFlags(flags types.Flags) uint32 {
	count := uint32(0)
	err := m.mailstore.Db.Get(&count,
		`SELECT COUNT(*)
		 FROM mail_messages
		 WHERE mailbox_id=?
			AND NOT flags & ? = ?`,
		m.ID,
		flags,
		flags,
	)
	if err != nil {
		panic(err)
	}
	return count
}

// Recalculate and write the sequence numbers to the database
// This should be called whenever a message is deleted or created in the mailbox
func (m MySQLMailstore) recalculateSequenceNumbers(mailboxID uint32, transaction sqlx.Tx) error {
	fmt.Println("recalculateSequenceNumbers not implemented")
	return nil
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
		Db:   db,
		User: nil,
	}

	ms.User = &MySQLUser{
		mailstore:     ms,
		authenticated: false,
	}
	return ms, err
}

// Authenticate implements the Authenticate method on the Mailstore interface
func (d MySQLMailstore) Authenticate(username string, password string) (User, error) {
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
func (u MySQLUser) Mailboxes() []Mailbox {
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
func (u MySQLUser) MailboxByName(name string) (Mailbox, error) {
	result := MySQLMailbox{}
	err := u.mailstore.Db.Get(&result,
		`SELECT id, name
		 FROM mail_mailboxes
		 WHERE name=?`, name)

	if err != nil {
		return result, err
	}

	if result.name == "" {
		return nil, errors.New("Invalid mailbox")
	}
	return result, nil
}

// MySQLMailbox is an in-memory implementation of a Mailstore Mailbox
type MySQLMailbox struct {
	ID        uint32 `db:"id"`
	name      string `db:"name"`
	nextuid   uint32 `db:"next_uid"`
	mailstore *MySQLMailstore
}

// Name returns the Mailbox's name
func (m MySQLMailbox) Name() string { return m.name }

// NextUID returns the UID that is likely to be assigned to the next
// new message in the Mailbox
func (m MySQLMailbox) NextUID() uint32 { return m.nextuid }

// LastUID returns the UID of the last message in the mailbox or if the
// mailbox is empty, the next expected UID
func (m MySQLMailbox) LastUID() uint32 {
	// TODO
	return 0
}

// Recent returns the number of messages in the mailbox which are currently
// marked with the 'Recent' flag
func (m MySQLMailbox) Recent() uint32 {
	return m.countMessagesWithFlags(types.FlagRecent)
}

// Messages returns the total number of messages in the Mailbox
func (m MySQLMailbox) Messages() uint32 {
	return m.countMessagesWithFlags(0)
}

// Unseen returns the number of messages in the mailbox which are currently
// marked with the 'Unseen' flag
func (m MySQLMailbox) Unseen() uint32 {
	return m.countMessagesWithoutFlags(types.FlagSeen)
}

// MessageBySequenceNumber returns a single message given the message's sequence number
func (m MySQLMailbox) MessageBySequenceNumber(seqno uint32) Message {
	return m.messageByColumn("sequence_number", seqno)
}

// MessageByUID returns a single message given the message's sequence number
func (m MySQLMailbox) MessageByUID(uidno uint32) Message {
	return m.messageByColumn("uid", uidno)
}

func (m MySQLMailbox) messageByColumn(column string, value uint32) Message {
	message := MySQLMessage{}
	m.mailstore.Db.Get(&message,
		`SELECT *
		 FROM mail_messages
		 WHERE mailbox_id=?
			AND ?=?`,
		m.ID, column, value)

	message.mailstore = m.mailstore

	return message
}

func (m MySQLMailbox) messageRangeByColumn(column string, seqRange types.SequenceRange) []Message {
	msgs := make([]Message, 0)
	// If Min is "*", meaning the last UID in the mailbox, Max should
	// always be Nil
	if seqRange.Min.Last() {
		// Return the last message in the mailbox
		msg := MySQLMessage{}
		m.mailstore.Db.Get(&msg,
			`SELECT *
			 FROM mail_messages
			 WHERE mailbox_id=?
			 LIMIT 1
			 ORDER BY ? DESC`,
			m.ID, column,
		)
		msgs = append(msgs, msg)
		return msgs
	}

	min, err := seqRange.Min.Value()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return msgs
	}

	// If no Max is specified, the sequence number must be fixed
	if seqRange.Max.Nil() {
		var uid uint32
		// Fetch specific message by sequence number
		uid, err = seqRange.Min.Value()
		msg := m.MessageByUID(uid)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return msgs
		}
		if msg != nil {
			msgs = append(msgs, msg)
		}
		return msgs
	}

	max, err := seqRange.Max.Value()
	if seqRange.Max.Last() {
		err = m.mailstore.Db.Select(&msgs,
			`SELECT *
			 FROM mail_messages
			 WHERE mailbox_id=?
				AND ? >= ?
			 ORDER BY ? DESC`,
			m.ID, column, min, column,
		)
	} else {
		err = m.mailstore.Db.Select(&msgs,
			`SELECT *
			 FROM mail_messages
			 WHERE mailbox_id=?
				AND ? >= ? AND ? <= ?
			 ORDER BY ? DESC`,
			m.ID,
			column, min,
			column, max,
			column,
		)
	}

	return msgs
}

// MessageSetByUID returns a slice of messages given a set of UID ranges.
// eg 1,5,9,28:140,190:*
func (m MySQLMailbox) MessageSetByUID(set types.SequenceSet) []Message {
	var msgs []Message

	// If the mailbox is empty, return empty array
	if m.Messages() == 0 {
		return msgs
	}

	for _, msgRange := range set {
		msgs = append(msgs, m.messageRangeByColumn("uid", msgRange)...)
	}

	return msgs
}

// MessageSetBySequenceNumber returns a slice of messages given a set of
// sequence number ranges
func (m MySQLMailbox) MessageSetBySequenceNumber(set types.SequenceSet) []Message {
	var msgs []Message

	// If the mailbox is empty, return empty array
	if m.Messages() == 0 {
		return msgs
	}

	// For each sequence range in the sequence set
	for _, msgRange := range set {
		msgs = append(msgs, m.messageRangeByColumn("sequence_number", msgRange)...)
	}
	return msgs

}

// NewMessage creates a new message in the dummy mailbox.
func (m MySQLMailbox) NewMessage() Message {
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
func (m MySQLMessage) Header() (hdr textproto.MIMEHeader) {
	return m.header
}

// UID returns the message's unique identifier (UID).
func (m MySQLMessage) UID() uint32 { return m.uid }

// SequenceNumber returns the message's sequence number.
func (m MySQLMessage) SequenceNumber() uint32 { return m.sequenceNumber }

// Size returns the message's full RFC822 size, including full message header
// and body.
func (m MySQLMessage) Size() uint32 {
	hdrStr := fmt.Sprintf("%s\r\n", m.Header())
	return uint32(len(hdrStr)) + uint32(len(m.Body()))
}

// InternalDate returns the internally stored date of the message
func (m MySQLMessage) InternalDate() time.Time {
	return m.internalDate
}

// Body returns the full body of the message
func (m MySQLMessage) Body() string {
	return m.body
}

// Keywords returns any keywords associated with the message
func (m MySQLMessage) Keywords() []string {
	var f []string
	//f[0] = "Test"
	return f
}

// Flags returns any flags on the message.
func (m MySQLMessage) Flags() types.Flags {
	return m.flags
}

// OverwriteFlags replaces any flags on the message with those specified.
func (m MySQLMessage) OverwriteFlags(newFlags types.Flags) Message {
	m.flags = newFlags
	return m
}

// AddFlags adds the given flag to the message.
func (m MySQLMessage) AddFlags(newFlags types.Flags) Message {
	m.flags = m.flags.SetFlags(newFlags)
	return m
}

// RemoveFlags removes the given flag from the message.
func (m MySQLMessage) RemoveFlags(newFlags types.Flags) Message {
	m.flags = m.flags.ResetFlags(newFlags)
	return m
}

// SetHeaders sets the e-mail headers of the message.
func (m MySQLMessage) SetHeaders(newHeader textproto.MIMEHeader) Message {
	m.header = newHeader
	return m
}

// SetBody sets the body of the message.
func (m MySQLMessage) SetBody(newBody string) Message {
	m.body = newBody
	return m
}

// Save saves the message to the mailbox it belongs to.
func (m MySQLMessage) Save() (Message, error) {
	subject := m.Header().Get("subject")
	if m.sequenceNumber == 0 {
		nextUID := 8498489
		_, err := m.mailstore.Db.Exec(
			`INSERT INTO mail_messages
			 (uid, mailbox_id, date, flags, subject, header, body)
			 VALUES
			 (?, ?, ?, ?, ?, ?, ?)`,
			nextUID, m.mailboxID, m.internalDate, m.flags,
			subject, m.header, m.body,
		)
		if err != nil {
			return m, err
		}
		nextUID++
	} else {
		// Message exists
		_, err := m.mailstore.Db.Exec(
			`UPDATE mail_messages
			 SET date=?, flags=?, subject=?, header=?, body=?`,
			m.internalDate, m.flags,
			subject, m.header, m.body,
		)
		if err != nil {
			return m, err
		}
	}
	return m, nil
}

// DeleteFlaggedMessages deletes messages marked with the Delete flag and
// returns them.
func (m MySQLMailbox) DeleteFlaggedMessages() ([]Message, error) {
	var delMsgs []Message

	_, err := m.mailstore.Db.Exec(
		`DELETE FROM mail_messages WHERE mailbox_id=?
			AND flags & ? = ?`,
		m.ID,
		types.FlagDeleted,
		types.FlagDeleted,
	)

	return delMsgs, err
}
