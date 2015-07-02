package conn_test

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net/textproto"
	"strings"

	"github.com/jordwest/imap-server/conn"
	"github.com/jordwest/imap-server/mailstore"
	"github.com/jordwest/mock-conn"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var mStore mailstore.DummyMailstore
var tConn *conn.Conn
var mockConn *mock_conn.Conn
var reader *textproto.Reader

func TestConn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Conn Suite")
}

func SendLine(request string) {
	fmt.Fprintf(mockConn.Client, "%s\n", request)
}

func SendBase64(request string) {
	encoder := base64.NewEncoder(base64.StdEncoding, mockConn.Client)
	fmt.Fprintf(encoder, "%s", request)
	// Must close the encoder when finished to flush any partial blocks.
	encoder.Close()
}

func ExpectResponse(expected string) {
	response, err := reader.ReadLine()
	expected = strings.TrimSpace(expected)
	response = strings.TrimSpace(response)
	Expect(response, err).To(Equal(expected))
}

// === SETUP ====
var _ = BeforeEach(func() {
	mStore = mailstore.NewDummyMailstore()
	mockConn = mock_conn.NewConn()
	tConn = conn.NewConn(mStore, mockConn.Server, GinkgoWriter)
	reader = textproto.NewReader(bufio.NewReader(mockConn.Client))
})

var _ = JustBeforeEach(func() {
	go tConn.Start()
})

// === TEARDOWN ====

var _ = AfterEach(func() {
	tConn.Close()
	mockConn.Close()
})
