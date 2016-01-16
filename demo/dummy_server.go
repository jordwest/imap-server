package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"

	imap "github.com/jordwest/imap-server"
	"github.com/jordwest/imap-server/mailstore"
)

var tlsCertificatePath = flag.String("tls_certificate_path", "", "Path to the TLS certificate, in PEM format. Specifying this enables STARTTLS. Specifying this requires specifying flag tls_privatekey_path, too. Generate certificates with: openssl req -nodes -new -x509 -keyout /path/to/private.key -out /path/to/certificate.crt")
var tlsPrivateKeyPath = flag.String("tls_privatekey_path", "", "Path to the TLS private key, in PEM format. This flag is ignored if flag tls_certificate_path is not specified.")

func init() {
	flag.Parse()
}

func main() {
	store := mailstore.NewDummyMailstore()
	s := imap.NewServer(store)
	s.Transcript = os.Stdout
	s.Addr = ":10143"

	if *tlsCertificatePath != "" {
		if *tlsPrivateKeyPath == "" {
			panic("You need to specify -tls_privatekey_path flag.")
		}
		cert, err := tls.LoadX509KeyPair(*tlsCertificatePath, *tlsPrivateKeyPath)
		if err != nil {
			panic("Failed to load keypair: " + err.Error())
		}

		s.StartTLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	err := s.ListenAndServe()
	if err != nil {
		fmt.Printf("Error creating test connection: %s\n", err)
	}
}
