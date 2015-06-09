package types

// Flags provide information on flags that are attached to a message
type Flags struct {
	seen bool
}

func (f Flags) IsSeen() bool {
	return f.seen
}
