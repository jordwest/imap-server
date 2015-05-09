package imap

import (
	"fmt"
	"strings"
)

type MIMEHeader map[string]string

// Perform case-insensitive search
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
