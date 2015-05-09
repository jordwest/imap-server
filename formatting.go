package imap

import (
	"fmt"
	"strings"
	"time"
)

// RFC822 date format used by IMAP in go date format
const rfc822Date = "Mon, 2 Jan 2006 15:04:05 +0700"

// Date format used in INTERNALDATE fetch parameter
const internalDate = "02-Jan-2006 15:04:05 +0700"

func formatDate(date time.Time) string {
	fmt.Printf("date: %s\n", date)
	return date.Format(rfc822Date)
}

func splitParams(params string) []string {
	paramsOpen := false
	result := strings.FieldsFunc(params, func(r rune) bool {
		if r == '[' {
			paramsOpen = true
		}
		if r == ']' {
			paramsOpen = false
		}
		if r == ' ' && !paramsOpen {
			return true
		}
		return false
	})
	return result
}

func debugPrintMessages(messages []Message) {
	fmt.Printf("SeqNo  |UID    |From      |To        |Subject\n")
	fmt.Printf("-------+-------+----------+----------+-------\n")
	for _, msg := range messages {
		_, from, _ := msg.Header().FindKey("from")
		_, to, _ := msg.Header().FindKey("to")
		_, subject, _ := msg.Header().FindKey("subject")
		fmt.Printf("%-7d|%-7d|%-10.10s|%-10.10s|%s\n", msg.SequenceNumber(), msg.UID(), from, to, subject)
	}
}
