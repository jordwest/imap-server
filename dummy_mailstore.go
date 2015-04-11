package imap_server

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
	}
}

func NewDummyMailstore() DummyMailstore {
	ms := DummyMailstore{
		user: DummyUser{
			authenticated: false,
			mailboxes:     make(map[string]DummyMailbox),
		},
	}
	ms.user.mailboxes["INBOX"] = newDummyMailbox("INBOX")
	ms.user.mailboxes["Trash"] = newDummyMailbox("Trash")
	return ms
}

func (d DummyMailstore) Authenticate(username string, password string) (User, error) {
	if username != "username" {
		return DummyUser{}, errors.New("Invalid username. Use 'username'")
	}

	if password != "password" {
		return DummyUser{}, errors.New("Invalid password. Use 'password'")
	}

	return DummyUser{authenticated: true}, nil
}

type DummyUser struct {
	authenticated bool
	mailboxes     map[string]DummyMailbox
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
	messages []Message
}

func (m DummyMailbox) Name() string   { return m.name }
func (m DummyMailbox) NextUid() int64 { return 2 }
func (m DummyMailbox) Recent() int    { return 1 }
func (m DummyMailbox) Messages() int  { return 1 }
func (m DummyMailbox) Unseen() int    { return 1 }

func (m DummyMailbox) MessageBySequenceNumber(seqno int) Message {
	return DummyMessage{
		sequenceNumber: seqno,
		uid:            seqno,
	}
}

func (m DummyMailbox) MessageByUid(uidno int) Message {
	return DummyMessage{
		sequenceNumber: uidno,
		uid:            uidno,
	}
}

func (m DummyMailbox) MessageRangeByUid(startUid int, endUid int) []Message {
	msgs := make([]Message, 2)
	msgs[0] = m.MessageByUid(1)
	msgs[1] = m.MessageByUid(2)
	return msgs

}

func (m DummyMailbox) MessageRangeBySequenceNumber(startUid int, endUid int) []Message {
	msgs := make([]Message, 2)
	msgs[0] = m.MessageBySequenceNumber(1)
	msgs[1] = m.MessageBySequenceNumber(2)
	return msgs

}

type DummyMessage struct {
	sequenceNumber int
	uid            int
}

func (m DummyMessage) Header() (hdr MIMEHeader) {
	hdr = make(map[string]string)
	hdr["Date"] = "Mon, 27 Oct 2014 13:45:00 +1000"
	hdr["To"] = "you@test.com"
	hdr["From"] = "me@test.com"
	hdr["Subject"] = "This is a dummy email"
	hdr["Message-ID"] = "<123456@test.com>"
	return hdr
}

func (m DummyMessage) Uid() int { return m.uid }
func (m DummyMessage) SequenceNumber() int {
	return m.sequenceNumber
}

func (m DummyMessage) Size() int {
	hdrStr := fmt.Sprintf("%s\r\n", m.Header())
	return len(hdrStr) + len(m.Body())
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

func (m DummyMessage) IsSeen() bool     { return false }
func (m DummyMessage) IsAnswered() bool { return false }
func (m DummyMessage) IsFlagged() bool  { return false }
func (m DummyMessage) IsDeleted() bool  { return false }
func (m DummyMessage) IsDraft() bool    { return false }
func (m DummyMessage) IsRecent() bool   { return true }
