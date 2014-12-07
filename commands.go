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

var commands []command

// Register all supported client command handlers
// with the server. This function is run on server startup and
// panics if a command regex is invalid.
func init() {
	commands = make([]command, 0)

	registerCommand("CAPABILITY", cmdCapability)
	registerCommand("LOGIN \"([A-z0-9]+)\" \"([A-z0-9]+)\"", cmdLogin)
	registerCommand("AUTHENTICATE PLAIN", cmdAuthPlain)
	registerCommand("LIST", cmdList)
	registerCommand("LSUB", cmdLSub)
	registerCommand("LOGOUT", cmdLogout)
	registerCommand("NOOP", cmdNoop)
	registerCommand("CLOSE", cmdClose)
	registerCommand("SELECT \"?([A-z0-9]+)\"?", cmdSelect)
	registerCommand("EXAMINE \"?([A-z0-9]+)\"?", cmdExamine)
	registerCommand("STATUS \"?([A-z0-9]+)\"? \\(([A-z\\s]+)\\)", cmdStatus)
	registerCommand("(UID )?FETCH (?:(\\d+)(?:\\:([\\*\\d]+))?) \\(([A-z0-9\\s\\(\\)\\[\\]\\.-]+)\\)", cmdFetch)
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
	c.writeResponse("", "LIST () \"/\" INBOX")
	c.writeResponse(args.Id(), "OK LIST Completed")
}

func cmdStatus(args commandArgs, c *Conn) {
	c.writeResponse("", "STATUS "+args.Arg(0)+" (UIDNEXT 2 UNSEEN 1)")
	c.writeResponse(args.Id(), "OK STATUS Completed")
}

func cmdLSub(args commandArgs, c *Conn) {
	c.writeResponse("", "LSUB () \"/\" INBOX")
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

func cmdFetch(args commandArgs, c *Conn) {
	start, _ := strconv.Atoi(args.Arg(1))

	searchByUid := args.Arg(0) == "UID "

	// Fetch the messages
	var msg Message
	if searchByUid {
		fmt.Printf("Searching by UID\n")
		msg = c.selectedMailbox.MessageByUid(start)
	} else {
		msg = c.selectedMailbox.MessageBySequenceNumber(start)
	}

	uidRE := regexp.MustCompile("UID")
	internalDateRE := regexp.MustCompile("INTERNALDATE")
	sizeRE := regexp.MustCompile("RFC822\\.SIZE")
	headersRE := regexp.MustCompile("BODY(?:\\.PEEK)?\\[HEADER\\]")
	specificHeadersRE := regexp.MustCompile("BODY(?:\\.PEEK)?" +
		"\\[HEADER\\.FIELDS \\(([A-z\\s-]+)\\)\\]")
	bodyOnlyRE := regexp.MustCompile("BODY(?:\\.PEEK)?\\[TEXT\\]")
	fullTextRE := regexp.MustCompile("BODY(?:\\.PEEK)?\\[\\]")
	peekRE := regexp.MustCompile("\\.PEEK")

	partsRequested := args.Arg(3)
	peekRequested := peekRE.MatchString(partsRequested)
	peekStr := ""
	if peekRequested {
		peekStr = ".PEEK"
	}

	reply := make([]string, 0)

	// Did the client request the message UID?
	if uidRE.MatchString(partsRequested) || searchByUid {
		reply = append(reply, fmt.Sprintf("UID %d", msg.Uid()))
	}

	// Did the client request all mail headers?
	if headersRE.MatchString(partsRequested) {
		hdr := fmt.Sprintf("\r\n%s\r\n\r\n", msg.Header())
		hdrLen := len(hdr)
		reply = append(reply, fmt.Sprintf("BODY%s[HEADER] {%d}%s", peekStr, hdrLen, hdr))
	}

	// Did the client request internal date?
	if internalDateRE.MatchString(partsRequested) {
		dateStr := formatDate(msg.InternalDate())
		reply = append(reply, fmt.Sprintf("INTERNALDATE \"%s\"", dateStr))
	}

	// Did the client request size?
	if sizeRE.MatchString(partsRequested) {
		reply = append(reply, fmt.Sprintf("RFC822.SIZE %d", msg.Size()))
	}

	// Did the client request only a specific subset of headers?
	specificHeadersRequested := specificHeadersRE.FindStringSubmatch(partsRequested)
	if len(specificHeadersRequested) > 0 {
		if !peekRequested {
			fmt.Printf("TODO: Peek not requested, mark all as non-recent\n")
		}
		fields := strings.Split(specificHeadersRequested[1], " ")
		hdrs := msg.Header()
		requestedHeaders := make(MIMEHeader)
		for _, key := range fields {
			// If the key exists in the headers, copy it over
			if k, v, ok := hdrs.FindKey(key); ok {
				requestedHeaders[k] = v
			}
		}
		hdr := fmt.Sprintf("\r\n%s\r\n\r\n", requestedHeaders)
		hdrLen := len(hdr)
		reply = append(reply, fmt.Sprintf("BODY%s[HEADER.FIELDS (%s)] {%d}%s",
			peekStr,
			specificHeadersRequested[1],
			hdrLen,
			hdr))
	}

	// Did the client request only the body text of the email?
	if bodyOnlyRE.MatchString(partsRequested) {
		body := fmt.Sprintf("\r\n%s\r\n", msg.Body())
		bodyLen := len(body)

		reply = append(reply, fmt.Sprintf("BODY%s[TEXT] {%d}%s",
			peekStr, bodyLen, body))
	}

	// Did the client request the complete email?
	fullTextRequested := fullTextRE.MatchString(partsRequested)
	if fullTextRequested {
		mail := fmt.Sprintf("\r\n%s\r\n\r\n%s\r\n", msg.Header(), msg.Body())
		mailLen := len(mail)

		reply = append(reply, fmt.Sprintf("BODY%s[] {%d}%s",
			peekStr, mailLen, mail))
	}

	fullReply := fmt.Sprintf("%d FETCH (%s)",
		msg.SequenceNumber(),
		strings.Join(reply, " "))

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
