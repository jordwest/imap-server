package types

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
