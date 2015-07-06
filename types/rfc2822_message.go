package types

import (
	"bufio"
	"bytes"
	"io"
	"net/textproto"
)

type RFC2822Message struct {
	Headers textproto.MIMEHeader
	Body    string
}

func MessageFromBytes(msgBytes []byte) (msg RFC2822Message, err error) {
	// The header and body are separated by a double new-line
	splitMessage := bytes.SplitN(msgBytes, []byte{'\r', '\n', '\r', '\n'}, 2)

	// Read the headers
	headerReader := textproto.NewReader(bufio.NewReader(bytes.NewReader(splitMessage[0])))
	msg.Headers, err = headerReader.ReadMIMEHeader()

	if err != nil && err != io.EOF {
		return msg, err
	}

	if len(splitMessage) == 2 {
		msg.Body = string(splitMessage[1])
	}

	return msg, nil
}
