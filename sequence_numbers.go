package imap_server

import (
	"fmt"
	"regexp"
	"strings"
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

type errInvalidRangeString string
type errInvalidSequenceSetString string

func (e errInvalidRangeString) Error() string {
	return fmt.Sprintf("Invalid sequence range string '%s' specified", string(e))
}
func (e errInvalidSequenceSetString) Error() string {
	return fmt.Sprintf("Invalid sequence set string '%s' specified", string(e))
}

var rangeRegexp *regexp.Regexp
var setRegexp *regexp.Regexp

func init() {
	// Regex for finding a sequence range
	rangeRegexp = regexp.MustCompile("^(\\d{1,10}|\\*)" + // Range lower bound - digit or star
		"(?:\\:(\\d{1,10}|\\*))?$") // Range upper bound - digit or star

	// Regex for finding a sequence-set - ie, multiple sequence ranges
	setRegexp = regexp.MustCompile("^((?:\\d{1,10}|\\*)(?:\\:(?:\\d{1,10}|\\*))?)" + // First range
		"(?:" + // Match zero or more additional ranges
		"," + // Must be separated by a comma
		"((?:\\d{1,10}|\\*)(?:\\:(?:\\d{1,10}|\\*))?)" + // Additional ranges
		")*" + // Match zero or more
		"$")
}

func interpretMessageRange(imapMessageRange string) (seqRange SequenceRange, err error) {
	result := rangeRegexp.FindStringSubmatch(imapMessageRange)
	if len(result) == 0 {
		return SequenceRange{}, errInvalidRangeString(imapMessageRange)
	}

	return SequenceRange{min: result[1], max: result[2]}, nil
}

func interpretSequenceSet(imapSequenceSet string) (seqSet SequenceSet, err error) {
	// Ensure the sequence set is valid
	if !setRegexp.MatchString(imapSequenceSet) {
		return SequenceSet{}, errInvalidSequenceSetString(imapSequenceSet)
	}

	ranges := strings.Split(imapSequenceSet, ",")

	seqSet = make(SequenceSet, len(ranges))
	for index, rng := range ranges {
		seqSet[index], err = interpretMessageRange(rng)
		if err != nil {
			return seqSet, err
		}
	}

	return seqSet, nil
}
