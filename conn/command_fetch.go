package conn

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jordwest/imap-server/mailstore"
	"github.com/jordwest/imap-server/types"
	"github.com/jordwest/imap-server/util"
)

var registeredFetchParams []fetchParamDefinition
var peekRE *regexp.Regexp

// ErrUnrecognisedParameter indicates that the parameter requested in a FETCH
// command is unrecognised or not implemented in this IMAP server
var ErrUnrecognisedParameter = errors.New("Unrecognised Parameter")

type fetchParamDefinition struct {
	re      *regexp.Regexp
	handler func([]string, *Conn, mailstore.Message, bool) string
}

func cmdFetch(args commandArgs, c *Conn) {
	start, _ := strconv.Atoi(args.Arg(1))

	searchByUID := args.Arg(0) == "UID "

	// Fetch the messages
	var msg mailstore.Message
	if searchByUID {
		fmt.Printf("Searching by UID\n")
		msg = c.SelectedMailbox.MessageByUID(uint32(start))
	} else {
		msg = c.SelectedMailbox.MessageBySequenceNumber(uint32(start))
	}

	fetchParamString := args.Arg(3)
	if searchByUID && !strings.Contains(fetchParamString, "UID") {
		fetchParamString += " UID"
	}

	fetchParams, err := fetch(fetchParamString, c, msg)
	if err != nil {
		if err == ErrUnrecognisedParameter {
			c.writeResponse(args.ID(), "BAD Unrecognised Parameter")
			return
		}

		c.writeResponse(args.ID(), "BAD")
		return
	}

	fullReply := fmt.Sprintf("%d FETCH (%s)",
		msg.SequenceNumber(),
		fetchParams)

	c.writeResponse("", fullReply)
	if searchByUID {
		c.writeResponse(args.ID(), "OK UID FETCH Completed")
	} else {
		c.writeResponse(args.ID(), "OK FETCH Completed")
	}
}

// Fetch requested params from a given message
// eg fetch("UID BODY[TEXT] RFC822.SIZE", c, message)
func fetch(params string, c *Conn, m mailstore.Message) (string, error) {
	paramList := util.SplitParams(params)

	// Prepare the list of responses
	responseParams := make([]string, 0, len(paramList))

	for _, param := range paramList {
		paramResponse, err := fetchParam(param, c, m)
		if err != nil {
			return "", err
		}
		responseParams = append(responseParams, paramResponse)
	}
	return strings.Join(responseParams, " "), nil
}

// Match a single fetch parameter and return the data
func fetchParam(param string, c *Conn, m mailstore.Message) (string, error) {
	peek := false
	if peekRE.MatchString(param) {
		peek = true
	}
	// Search through the parameter list until a parameter handler is found
	for _, element := range registeredFetchParams {
		if element.re.MatchString(param) {
			return element.handler(element.re.FindStringSubmatch(param), c, m, peek), nil
		}
	}
	return "", ErrUnrecognisedParameter
}

// Register all supported fetch parameters
func init() {
	peekRE = regexp.MustCompile("\\.PEEK")
	registeredFetchParams = make([]fetchParamDefinition, 0)
	registerFetchParam("UID", fetchUID)
	registerFetchParam("FLAGS", fetchFlags)
	registerFetchParam("RFC822\\.SIZE", fetchRfcSize)
	registerFetchParam("INTERNALDATE", fetchInternalDate)
	registerFetchParam("BODY(?:\\.PEEK)?\\[HEADER\\]", fetchHeaders)
	registerFetchParam("BODY(?:\\.PEEK)?"+
		"\\[HEADER\\.FIELDS \\(([A-z\\s-]+)\\)\\]", fetchHeaderSpecificFields)
	registerFetchParam("BODY(?:\\.PEEK)?\\[TEXT\\]", fetchBody)
	registerFetchParam("BODY(?:\\.PEEK)?\\[\\]", fetchFullText)
}

func registerFetchParam(regex string, handler func([]string, *Conn, mailstore.Message, bool) string) {
	newParam := fetchParamDefinition{
		re:      regexp.MustCompile(regex),
		handler: handler,
	}
	registeredFetchParams = append(registeredFetchParams, newParam)
}

// Fetch the UID of the mail message
func fetchUID(args []string, c *Conn, m mailstore.Message, peekOnly bool) string {
	return fmt.Sprintf("UID %d", m.UID())
}

func fetchFlags(args []string, c *Conn, m mailstore.Message, peekOnly bool) string {
	flags := append(m.Flags().Strings(), m.Keywords()...)
	flagList := strings.Join(flags, " ")
	return fmt.Sprintf("FLAGS (%s)", flagList)
}

func fetchRfcSize(args []string, c *Conn, m mailstore.Message, peekOnly bool) string {
	return fmt.Sprintf("RFC822.SIZE %d", m.Size())
}

func fetchInternalDate(args []string, c *Conn, m mailstore.Message, peekOnly bool) string {
	dateStr := m.InternalDate().Format(util.InternalDate)
	return fmt.Sprintf("INTERNALDATE \"%s\"", dateStr)
}

func fetchHeaders(args []string, c *Conn, m mailstore.Message, peekOnly bool) string {
	hdr := fmt.Sprintf("\r\n%s\r\n\r\n", m.Header())
	hdrLen := len(hdr)

	peekStr := ""
	if peekOnly {
		peekStr = ".PEEK"
	}

	return fmt.Sprintf("BODY%s[HEADER] {%d}%s", peekStr, hdrLen, hdr)
}

func fetchHeaderSpecificFields(args []string, c *Conn, m mailstore.Message, peekOnly bool) string {
	if !peekOnly {
		fmt.Printf("TODO: Peek not requested, mark all as non-recent\n")
	}
	fields := strings.Split(args[1], " ")
	hdrs := m.Header()
	requestedHeaders := make(types.MIMEHeader)
	replyFieldList := make([]string, len(fields))
	for i, key := range fields {
		replyFieldList[i] = "\"" + key + "\""
		// If the key exists in the headers, copy it over
		if k, v, ok := hdrs.FindKey(key); ok {
			requestedHeaders[k] = v
		}
	}
	hdr := fmt.Sprintf("\r\n%s\r\n\r\n", requestedHeaders)
	hdrLen := len(hdr)

	return fmt.Sprintf("BODY[HEADER.FIELDS (%s)] {%d}%s",
		strings.Join(replyFieldList, " "),
		hdrLen,
		hdr)

}

func fetchBody(args []string, c *Conn, m mailstore.Message, peekOnly bool) string {
	body := fmt.Sprintf("\r\n%s\r\n", m.Body())
	bodyLen := len(body)

	return fmt.Sprintf("BODY[TEXT] {%d}%s",
		bodyLen, body)
}

func fetchFullText(args []string, c *Conn, m mailstore.Message, peekOnly bool) string {
	mail := fmt.Sprintf("\r\n%s\r\n\r\n%s\r\n", m.Header(), m.Body())
	mailLen := len(mail)

	return fmt.Sprintf("BODY[] {%d}%s",
		mailLen, mail)
}
