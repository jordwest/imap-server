package imap_server

import (
	"errors"
	"regexp"
)

var errInvalidRangeString = errors.New("Invalid message identifier range specified")

var rangeRegexp *regexp.Regexp

func init() {
	rangeRegexp = regexp.MustCompile("^(\\d{1,10}|\\*)(?:\\:(\\d{1,10}|\\*))?$")
}

func interpretMessageRange(imapMessageRange string) (min string, max string, err error) {
	result := rangeRegexp.FindStringSubmatch(imapMessageRange)
	if len(result) == 0 {
		return "", "", errInvalidRangeString
	}

	return result[1], result[2], nil
}
