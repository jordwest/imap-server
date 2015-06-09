package types

import "strings"

// Flags provide information on flags that are attached to a message
type Flags int32

const (
	FlagSeen Flags = 1 << iota
	FlagAnswered
	FlagFlagged
	FlagDeleted
	FlagDraft
	FlagRecent // Can not be set by client!
)

func CombineFlags(flags ...Flags) Flags {
	returnFlags := Flags(0)
	for _, f := range flags {
		returnFlags |= f
	}
	return returnFlags
}

func FlagsFromString(imapFlagString string) (f Flags) {
	for _, flag := range strings.Split(imapFlagString, " ") {
		switch flag {
		case "\\Seen":
			f = f.SetFlags(FlagSeen)
		case "\\Answered":
			f = f.SetFlags(FlagAnswered)
		case "\\Flagged":
			f = f.SetFlags(FlagFlagged)
		case "\\Deleted":
			f = f.SetFlags(FlagDeleted)
		case "\\Draft":
			f = f.SetFlags(FlagDraft)
		case "\\Recent":
			f = f.SetFlags(FlagRecent)
		}
	}
	return f
}

func (f Flags) ResetFlags(remove Flags) Flags {
	f ^= remove
	return f
}

func (f Flags) SetFlags(add Flags) Flags {
	f |= add
	return f
}

func (f Flags) HasFlags(check Flags) bool {
	return (f & check) == check
}

// Convert flags to list of IMAP format flags
func (f Flags) Strings() []string {
	flags := make([]string, 0, 6) // Up to 6 flags
	if f.HasFlags(FlagAnswered) {
		flags = append(flags, "\\Answered")
	}
	if f.HasFlags(FlagSeen) {
		flags = append(flags, "\\Seen")
	}
	if f.HasFlags(FlagRecent) {
		flags = append(flags, "\\Recent")
	}
	if f.HasFlags(FlagDeleted) {
		flags = append(flags, "\\Deleted")
	}
	if f.HasFlags(FlagDraft) {
		flags = append(flags, "\\Draft")
	}
	if f.HasFlags(FlagFlagged) {
		flags = append(flags, "\\Flagged")
	}
	return flags
}

// Convert flags to IMAP flag list string
func (f Flags) String() string {
	return strings.Join(f.Strings(), " ")
}
