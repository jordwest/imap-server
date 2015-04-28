package imap_server

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type command struct {
	match   *regexp.Regexp
	handler func(commandArgs, *Conn)
}

type commandArgs []string

func (a commandArgs) FullCommand() string {
	return a[0]
}

func (a commandArgs) Id() string {
	return a[1]
}

func (a commandArgs) Arg(i int) string {
	return a[i+2]
}

const LIST_ARG_SELECTOR int = 1

const STORE_ARG_OPERATION int = 3
const STORE_ARG_SILENT int = 4
const STORE_ARG_FLAGS int = 5

var commands []command

// Register all supported client command handlers
// with the server. This function is run on server startup and
// panics if a command regex is invalid.
func init() {
	commands = make([]command, 0)

	registerCommand("(?i:CAPABILITY)", cmdCapability)
	registerCommand("(?i:LOGIN) \"([A-z0-9]+)\" \"([A-z0-9]+)\"", cmdLogin)
	registerCommand("(?i:AUTHENTICATE PLAIN)", cmdAuthPlain)
	registerCommand("(?i:LIST) \"?([A-z0-9]+)?\"? \"?([A-z0-9*]+)?\"?", cmdList)
	registerCommand("(?i:LSUB)", cmdLSub)
	registerCommand("(?i:LOGOUT)", cmdLogout)
	registerCommand("(?i:NOOP)", cmdNoop)
	registerCommand("(?i:CLOSE)", cmdClose)
	registerCommand("(?i:SELECT) \"?([A-z0-9]+)?\"?", cmdSelect)
	registerCommand("(?i:EXAMINE) \"?([A-z0-9]+)\"?", cmdExamine)
	registerCommand("(?i:STATUS) \"?([A-z0-9/]+)\"? \\(([A-z\\s]+)\\)", cmdStatus)
	registerCommand("((?i)UID )?(?i:FETCH) (?:(\\d+)(?:\\:([\\*\\d]+))?) \\(([A-z0-9\\s\\(\\)\\[\\]\\.-]+)\\)", cmdFetch)
	// STORE 2:4 +FLAGS (\Deleted)       Mark messages as deleted
	// STORE 2:4 -FLAGS (\Seen)          Mark messages as unseen
	// STORE 2:4 FLAGS (\Seen \Deleted)  Replace flags
	registerCommand("((?i)UID )?(?i:STORE) (?:(\\d+)(?:\\:([\\*\\d]+))?) ([\\+\\-])?(?i:FLAGS(\\.SILENT)?) \\(?([\\\\A-z0-9]+)\\)?", cmdStoreFlags)
}

func registerCommand(matchExpr string, handleFunc func(commandArgs, *Conn)) error {
	// Add command identifier to beginning of command
	matchExpr = "([A-z0-9\\.]+) " + matchExpr

	newRE := regexp.MustCompile(matchExpr)
	c := command{match: newRE, handler: handleFunc}
	commands = append(commands, c)
	return nil
}

// Handles a CAPABILITY command
func cmdCapability(args commandArgs, c *Conn) {
	c.writeResponse("", "CAPABILITY IMAP4rev1 AUTH=PLAIN")
	c.writeResponse(args.Id(), "OK CAPABILITY completed")
}

// Handles PLAIN text AUTHENTICATE command
func cmdAuthPlain(args commandArgs, c *Conn) {
	// Compile login regex
	loginRE := regexp.MustCompile("(?:[A-z0-9]+)\x00([A-z0-9]+)\x00([A-z0-9]+)")

	// Tell client to go ahead
	c.writeResponse("+", "")

	// Wait for client to send auth details
	authDetails := <-c.recvReq
	data, err := base64.StdEncoding.DecodeString(authDetails)
	if err != nil {
		c.writeResponse("", "BAD Invalid auth details")
		return
	}
	fmt.Printf("Auth details received: %q\n", data)
	match := loginRE.FindSubmatch(data)
	if len(match) != 3 {
		c.writeResponse(args.Id(), "NO Incorrect username/password")
		return
	}
	c.user, err = c.srv.mailstore.Authenticate(string(match[1]), string(match[2]))
	if err != nil {
		c.writeResponse(args.Id(), "NO Incorrect username/password")
		return
	}
	c.setState(StateAuthenticated)
	c.writeResponse(args.Id(), "OK Authenticated")
}

// Handles PLAIN text LOGIN command
func cmdLogin(args commandArgs, c *Conn) {
	user, err := c.srv.mailstore.Authenticate(args.Arg(0), args.Arg(1))
	c.user = user
	if err != nil {
		c.writeResponse(args.Id(), "NO Incorrect username/password")
		return
	}
	c.setState(StateAuthenticated)
	c.writeResponse(args.Id(), "OK Authenticated")
}

func cmdList(args commandArgs, c *Conn) {
	if args.Arg(LIST_ARG_SELECTOR) == "" {
		// Blank selector means request directory separator
		c.writeResponse("", "LIST (\\Noselect) \"/\" \"\"")
	} else if args.Arg(LIST_ARG_SELECTOR) == "*" {
		// List all mailboxes requested
		for _, mailbox := range c.user.Mailboxes() {
			c.writeResponse("", "LIST () \"/\" \""+mailbox.Name()+"\"")
		}
	}
	c.writeResponse(args.Id(), "OK LIST completed")
}

func cmdStatus(args commandArgs, c *Conn) {
	mailbox, err := c.user.MailboxByName(args.Arg(0))
	if err != nil {
		c.writeResponse(args.Id(), "NO "+err.Error())
		return
	}

	c.writeResponse("", fmt.Sprintf("STATUS %s (UIDNEXT %d UNSEEN %d)",
		mailbox.Name(), mailbox.NextUid(), mailbox.Unseen()))
	c.writeResponse(args.Id(), "OK STATUS Completed")
}

func cmdLSub(args commandArgs, c *Conn) {
	for _, mailbox := range c.user.Mailboxes() {
		c.writeResponse("", "LSUB () \"/\" \""+mailbox.Name()+"\"")
	}
	c.writeResponse(args.Id(), "OK LSUB Completed")
}

func cmdLogout(args commandArgs, c *Conn) {
	c.writeResponse("", "BYE IMAP4rev1 server logging out")
	c.setState(StateLoggedOut)
	c.writeResponse(args.Id(), "OK LOGOUT completed")
	c.Close()
}

// Write out the info for a mailbox (used in both SELECT and EXAMINE)
func writeMailboxInfo(c *Conn, m Mailbox) {
	fmt.Fprintf(c, "* %d EXISTS\r\n", m.Messages())
	fmt.Fprintf(c, "* %d RECENT\r\n", m.Recent())
	fmt.Fprintf(c, "* OK [UNSEEN %d]\r\n", m.Unseen())
	fmt.Fprintf(c, "* OK [UIDNEXT %d]\r\n", m.NextUid())
	fmt.Fprintf(c, "* OK [UIDVALIDITY %d]\r\n", 250)
	fmt.Fprintf(c, "* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)\r\n")
}

func cmdSelect(args commandArgs, c *Conn) {
	var err error
	c.selectedMailbox, err = c.user.MailboxByName(args.Arg(0))
	if err != nil {
		fmt.Fprintf(c, "%s NO %s\r\n", args.Id(), err)
		return
	}

	writeMailboxInfo(c, c.selectedMailbox)
	c.writeResponse(args.Id(), "OK [READ-WRITE] SELECT completed")
}

func cmdExamine(args commandArgs, c *Conn) {
	m, err := c.user.MailboxByName(args.Arg(0))
	if err != nil {
		fmt.Fprintf(c, "%s NO %s\r\n", args.Id(), err)
		return
	}

	writeMailboxInfo(c, m)
	c.writeResponse(args.Id(), "OK [READ-ONLY] EXAMINE completed")
}

func messageFlags(msg Message) []string {
	flags := make([]string, 0)
	if msg.IsAnswered() {
		flags = append(flags, "\\Answered")
	}
	if msg.IsSeen() {
		flags = append(flags, "\\Seen")
	}
	if msg.IsRecent() {
		flags = append(flags, "\\Recent")
	}
	if msg.IsDeleted() {
		flags = append(flags, "\\Deleted")
	}
	if msg.IsDraft() {
		flags = append(flags, "\\Draft")
	}
	if msg.IsFlagged() {
		flags = append(flags, "\\Flagged")
	}
	return flags
}

func cmdStoreFlags(args commandArgs, c *Conn) {
	fmt.Printf("STORE command args: %+v\n\n", args)
	operation := args.Arg(STORE_ARG_OPERATION)
	flags := args.Arg(STORE_ARG_FLAGS)

	silent := false
	if args.Arg(STORE_ARG_SILENT) == ".SILENT" {
		silent = true
	}
	if silent {
		fmt.Printf("Silently ")
	}

	if operation == "+" {
		fmt.Printf("Add flags %s\n", flags)
	} else if operation == "-" {
		fmt.Printf("Remove flags %s\n", flags)
	} else {
		fmt.Printf("Set flags %s\n", flags)
	}

	c.writeResponse(args.Id(), "OK STORE Completed")
}

func cmdFetch(args commandArgs, c *Conn) {
	start, _ := strconv.Atoi(args.Arg(1))

	searchByUid := args.Arg(0) == "UID "

	// Fetch the messages
	var msg Message
	if searchByUid {
		fmt.Printf("Searching by UID\n")
		msg = c.selectedMailbox.MessageByUid(int32(start))
	} else {
		msg = c.selectedMailbox.MessageBySequenceNumber(start)
	}

	fetchParamString := args.Arg(3)
	if searchByUid && !strings.Contains(fetchParamString, "UID") {
		fetchParamString += " UID"
	}

	fetchParams, err := fetch(fetchParamString, c, msg)
	if err != nil {
		if err == UnrecognisedParameterError {
			c.writeResponse(args.Id(), "BAD Unrecognised Parameter")
			return
		} else {
			c.writeResponse(args.Id(), "BAD")
			return
		}
	}

	fullReply := fmt.Sprintf("%d FETCH (%s)",
		msg.SequenceNumber(),
		fetchParams)

	c.writeResponse("", fullReply)
	if searchByUid {
		c.writeResponse(args.Id(), "OK UID FETCH Completed")
	} else {
		c.writeResponse(args.Id(), "OK FETCH Completed")
	}
}

func cmdNoop(args commandArgs, c *Conn) {
	c.writeResponse(args.Id(), "OK NOOP Completed")
}

func cmdClose(args commandArgs, c *Conn) {
	c.setState(StateAuthenticated)
	c.selectedMailbox = nil
	c.writeResponse(args.Id(), "OK CLOSE Completed")
}

func cmdNA(args commandArgs, c *Conn) {
	c.writeResponse(args.Id(), "BAD Not implemented")
}
