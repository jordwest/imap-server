package conn_test

import (
	"badc0de.net/pkg/dummycertbuilder"
	"crypto/tls"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("STARTTLS Command", func() {

	Context("When insecure", func() {
		BeforeEach(func() {
			tConn.Secure = false
			tConn.StartTLSConfig = &tls.Config{
				Certificates: []tls.Certificate{dummycertbuilder.GetDummyCert()},
			}
		})

		It("should switch to secure mode", func() {

			// TODO(ivucica): Fix SendLine/ExpectResponse to not break when we change the Rwc.
			//SendLine("abcd.123 STARTTLS")
			//ExpectResponse("abcd.123 OK STARTTLS starting.")

			// TODO(ivucica): What is the proper way fail the test if tConn.Secure != true?
			// TODO(ivucica): What is the proper way to fail the test if tConn is not an instance of crypto/tls.Server?
		})
	})

	Context("When secure", func() {
		BeforeEach(func() {
			tConn.Secure = true
		})

		It("should prevent switching to secure mode", func() {
			// TODO(ivucica): Fix SendLine/ExpectResponse to not break when we change the Rwc.
			//SendLine("abcd.123 STARTTLS")
			//ExpectResponse("abcd.123 BAD Already secure.")

		})
	})
})
