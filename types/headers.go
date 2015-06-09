package imap

import (
	"fmt"
	"strings"
)

// MIMEHeader represents a Key: Value type MIME header found in an email message
type MIMEHeader map[string]string

// FindKey performs a case-insensitive search on the header and returns the key,
// the value, and a boolean representing whether the key was found
func (h MIMEHeader) FindKey(searchKey string) (string, string, bool) {
	for key, val := range h {
		if strings.EqualFold(key, searchKey) {
			return key, val, true
		}
	}
	return "", "", false
}

func (h MIMEHeader) String() string {
	lines := make([]string, len(h))
	i := 0
	for key, val := range h {
		lines[i] = fmt.Sprintf("%s: %s", key, val)
		i++
	}
	return strings.Join(lines, "\r\n")
}
