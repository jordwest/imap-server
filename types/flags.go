package types

import "strings"

// Flags provide information on flags that are attached to a message
type Flags int32

// Various flags that are permitted to be set/unset.
const (
	FlagSeen Flags = 1 << iota
	FlagAnswered
	FlagFlagged
	FlagDeleted
	FlagDraft
	FlagRecent // Can not be set by client!
)

// CombineFlags meshes several flags into one so that all of them are set.
func CombineFlags(flags ...Flags) Flags {
	returnFlags := Flags(0)
	for _, f := range flags {
		returnFlags |= f
	}
	return returnFlags
}

// FlagsFromString returns the flags based on the input IMAP format string.
func FlagsFromString(imapFlagString string) Flags {
	var f Flags

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

// ResetFlags removes all set flags.
func (f Flags) ResetFlags(remove Flags) Flags {
	f &^= remove
	return f
}

// SetFlags sets the given flags.
func (f Flags) SetFlags(add Flags) Flags {
	f |= add
	return f
}

// HasFlags checks if the given flags are set.
func (f Flags) HasFlags(check Flags) bool {
	return (f & check) == check
}

// Strings convert flags to list of IMAP format flags.
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

// String converts flags to IMAP flag list string.
func (f Flags) String() string {
	return strings.Join(f.Strings(), " ")
}
