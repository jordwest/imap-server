package conn_test

import (
	"github.com/jordwest/imap-server/conn"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("LOGIN Command", func() {

	Context("When logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateAuthenticated)
			tConn.User = mStore.User
		})
		It("should return Authenticated", func() {
			SendLine(`abcd.123 login "username" "password"`)
			ExpectResponse("abcd.123 OK Authenticated")
		})
		It("should return incorrect user/pw", func() {
			SendLine(`abcd.123 login "username" "foo"`)
			ExpectResponse("abcd.123 NO Incorrect username/password")
		})
	})

	Context("When not logged in", func() {
		BeforeEach(func() {
			tConn.SetState(conn.StateNotAuthenticated)
		})
		It("should return Authenticated", func() {
			SendLine(`abcd.123 login "username" "password"`)
			ExpectResponse("abcd.123 OK Authenticated")
		})
		It("should return incorrect user/pw", func() {
			SendLine(`abcd.123 login "username" "foo"`)
			ExpectResponse("abcd.123 NO Incorrect username/password")
		})
	})
})
