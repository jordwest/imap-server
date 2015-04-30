package imap_server

import (
	"errors"
	"regexp"
)

// A single message identifier. Could be UID or sequence
// number. eg in IMAP: 8 or *
// See RFC3501 section 9
type SequenceNumber string

// A range of identifiers. eg in IMAP: 5:9 or 15:*
type SequenceRange struct {
	min string
	max string
}

// A set of sequence ranges. eg in IMAP: 1,3,5:9,18:*
type SequenceSet []SequenceRange

var errInvalidRangeString = errors.New("Invalid message identifier range specified")

var rangeRegexp *regexp.Regexp

func init() {
	rangeRegexp = regexp.MustCompile("^(\\d{1,10}|\\*)(?:\\:(\\d{1,10}|\\*))?$")
}

func interpretMessageRange(imapMessageRange string) (seqRange SequenceRange, err error) {
	result := rangeRegexp.FindStringSubmatch(imapMessageRange)
	if len(result) == 0 {
		return SequenceRange{}, errInvalidRangeString
	}

	return SequenceRange{min: result[1], max: result[2]}, nil
}
