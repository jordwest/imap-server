package imap_server

import (
	"fmt"
	"regexp"
	"strings"
)

var registeredFetchParams []fetchParamDefinition

type fetchParamDefinition struct {
	re      *regexp.Regexp
	handler func([]string, *Conn, Message, bool) string
}

// Fetch requested params from a given message
// eg fetch("UID BODY[TEXT] RFC822.SIZE", c, message)
func fetch(params string, c *Conn, m Message) {
	for _, def := range registeredFetchParams {
		args := def.re.FindStringSubmatch(params)
		if len(args) > 0 {

		}
	}
}

// Register all supported fetch parameters
func init() {
	registeredFetchParams = make([]fetchParamDefinition, 0)
	registerFetchParam("UID", fetchUid)
	registerFetchParam("FLAGS", fetchFlags)
	/*
		registerFetchParam("RFC822\\.SIZE", fetchRfcSize)
		registerFetchParam("BODY(?:\\.PEEK)?\\[HEADER\\]", bodyHeader)
		registerFetchParam("BODY(?:\\.PEEK)?"+
			"\\[HEADER\\.FIELDS \\(([A-z\\s-]+)\\)\\]", bodyHeaderSpecificFields)
		registerFetchParam("BODY(?:\\.PEEK)?\\[TEXT\\]", bodyHeader)
		registerFetchParam("BODY(?:\\.PEEK)?\\[\\]", bodyHeader)
	*/
}

func registerFetchParam(regex string, handler func([]string, *Conn, Message, bool) string) {
	newParam := fetchParamDefinition{
		re:      regexp.MustCompile(regex),
		handler: handler,
	}
	registeredFetchParams = append(registeredFetchParams, newParam)
}

// Fetch the UID of the mail message
func fetchUid(args []string, c *Conn, m Message, peekOnly bool) string {
	return fmt.Sprintf("UID %d", m.Uid())
}

func fetchFlags(args []string, c *Conn, m Message, peekOnly bool) string {
	flags := append(messageFlags(m), m.Keywords()...)
	flagList := strings.Join(flags, " ")
	return fmt.Sprintf("FLAGS (%s)", flagList)
}
