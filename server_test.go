package imap_server

import (
	"os"
	"testing"
)

type expectTest struct {
	testName          string
	injectState       func(*Conn)
	commands          []string
	expectedResponses []string
}

func state(state ConnState) func(server *Conn) {
	return func(server *Conn) {
		server.setState(state)
	}
}

var expectTests = []expectTest{
	{"Should send welcome message on connection init", nil, nil, []string{"* OK IMAP4rev1 Service Ready"}},

	{"Should send capabilities", state(StateNotAuthenticated),
		[]string{"abcd CAPABILITY"},
		[]string{"* CAPABILITY IMAP4rev1 AUTH=PLAIN",
			"abcd OK CAPABILITY completed"}},

	{"Should login", state(StateNotAuthenticated),
		[]string{"efgh LOGIN test test"},
		[]string{"efgh OK LOGIN COMPLETED"}},

	{"Should not login with incorrect password", state(StateNotAuthenticated),
		[]string{"efgh LOGIN test bad"},
		[]string{"efgh NO Incorrect username or password"}},
	/*
		**** TLS not implemented yet
		{"Should send capabilities when not in tls", state(StateNotAuthenticated),
			[]string{"abcd CAPABILITY"},
			[]string{"* CAPABILITY IMAP4rev1 STARTTLS LOGINDISABLED",
				"abcd OK CAPABILITY completed"}},

		{"Should send capabilities when in TLS mode", state(StateNotAuthenticated),
			[]string{"abcd CAPABILITY"},
			[]string{"* CAPABILITY IMAP4rev1 AUTH=PLAIN",
				"abcd OK CAPABILITY completed"}},
	*/
}

// Inject state into the server, send a command, then expect a particular response
func TestExpectedResponses(t *testing.T) {
	for _, test := range expectTests {
		func(test expectTest) {
			_, cConn, sConn, server, err := NewTestConnection()
			defer server.Close()

			sConn.Transcript = os.Stdout
			if err != nil {
				t.Errorf("Error creating test connection: %s", err)
				return
			}
			if test.injectState != nil {
				test.injectState(sConn)
			}
			go sConn.Start()
			if test.commands != nil {
				for _, cmd := range test.commands {
					cConn.PrintfLine("%s", cmd)
				}
			}
			for _, expectedResponse := range test.expectedResponses {
				line, err := cConn.ReadLine()
				if err != nil {
					t.Errorf("Error reading line: %s", err)
					return
				}
				if line != expectedResponse {
					t.Errorf("Actual response did not match expected:\n%s", expectedResponse)
					return
				}
			}
		}(test)
	}
}
