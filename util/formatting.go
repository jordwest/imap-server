package util

import (
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"strings"
	"time"
)

// RFC822 date format used by IMAP in go date format
const RFC822Date = "Mon, 2 Jan 2006 15:04:05 +0700"

// Date format used in INTERNALDATE fetch parameter
const InternalDate = "02-Jan-2006 15:04:05 +0700"

// FormatDate formats the given date in the RFC822 format.
func FormatDate(date time.Time) string {
	return date.Format(RFC822Date)
}

// SplitParams splits parameters in IMAP arguments so that they're easily
// readable.
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

// WriteMIMEHeader writes the MIME header out in the standard format. This
// should eventually be superseded by textproto.MIMEHeader.Write(w) once
// it is implemented in the go standard library.
func WriteMIMEHeader(writer io.Writer, header textproto.MIMEHeader) (n int, err error) {
	for k, vv := range header {
		for _, v := range vv {
			bytes, err := fmt.Fprintf(writer, "%s: %s\r\n", k, v)
			if err != nil {
				return n, err
			}
			n += bytes
		}
	}
	return n, nil
}

// MIMEHeaderToString converts a textproto.MIMEHeader into its string
// representation.
func MIMEHeaderToString(header textproto.MIMEHeader) string {
	buf := &bytes.Buffer{}
	_, err := WriteMIMEHeader(buf, header)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
