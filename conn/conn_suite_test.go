package conn_test

import (
	"bufio"
	"fmt"
	"net/textproto"

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

func ExpectResponse(response string) {
	Expect(reader.ReadLine()).To(Equal(response))
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
