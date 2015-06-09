package util

import (
	"fmt"
	"strings"
	"time"
)

// RFC822 date format used by IMAP in go date format
const RFC822Date = "Mon, 2 Jan 2006 15:04:05 +0700"

// Date format used in INTERNALDATE fetch parameter
const InternalDate = "02-Jan-2006 15:04:05 +0700"

func FormatDate(date time.Time) string {
	fmt.Printf("date: %s\n", date)
	return date.Format(RFC822Date)
}

func SplitParams(params string) []string {
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
