package imap_server

import (
	"fmt"
	"time"
)

const rfc822Date = "Mon, 2 Jan 2006 15:04:05 +0700"

func formatDate(date time.Time) string {
	fmt.Printf("date: %s\n", date)
	return date.Format(rfc822Date)
}
