Go IMAP Server
==============

Barebones IMAP4rev1 server for golang.

Supported Commands
------------------
Command       | Planned | Implemented | Tests
------------- | ------- | ----------- | -----
CAPABILITY    | ✓       | ✗           | ✗
NOOP          | ✓       | ✗           | ✗
LOGOUT        | ✓       | ✗           | ✗
AUTHENTICATE  | ✗       | -           | -
LOGIN         | ✓       | ✗           | ✗
STARTTLS      | ✓       | ✗           | ✗
EXAMINE       | ✓       | ✗           | ✗
CREATE        | ✗       | -           | -
DELETE        | ✗       | -           | -
RENAME        | ✗       | -           | -
SUBSCRIBE     | ✗       | -           | -
UNSUBSCRIBE   | ✗       | -           | -
LIST          | ✓       | ✗           | ✗
LSUB          | ✗       | -           | -
STATUS        | ✓       | ✗           | ✗
APPEND        | ✓       | ✗           | ✗
CHECK         | ?       | ✗           | ✗
CLOSE         | ✓       | ✗           | ✗
EXPUNGE       | ✓       | ✗           | ✗
SEARCH        | ✓       | ✗           | ✗
FETCH         | ✓       | ✗           | ✗
STORE         | ✓       | ✗           | ✗
COPY          | ✓       | ✗           | ✗
UID           | ✓       | ✗           | ✗
