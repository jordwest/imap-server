package imap

import (
	"fmt"
	"regexp"
)

type command struct {
	match   *regexp.Regexp
	handler func(commandArgs, *Conn)
}

type commandArgs []string

func (a commandArgs) FullCommand() string {
	return a[0]
}

func (a commandArgs) ID() string {
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

// Write out the info for a mailbox (used in both SELECT and EXAMINE)
func writeMailboxInfo(c *Conn, m Mailbox) {
	fmt.Fprintf(c, "* %d EXISTS\r\n", m.Messages())
	fmt.Fprintf(c, "* %d RECENT\r\n", m.Recent())
	fmt.Fprintf(c, "* OK [UNSEEN %d]\r\n", m.Unseen())
	fmt.Fprintf(c, "* OK [UIDNEXT %d]\r\n", m.NextUid())
	fmt.Fprintf(c, "* OK [UIDVALIDITY %d]\r\n", 250)
	fmt.Fprintf(c, "* FLAGS (\\Answered \\Flagged \\Deleted \\Seen \\Draft)\r\n")
}

func messageFlags(msg Message) []string {
	flags := make([]string, 0, 6) // Up to 6 flags
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

func cmdNA(args commandArgs, c *Conn) {
	c.writeResponse(args.ID(), "BAD Not implemented")
}
