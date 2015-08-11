![go-IMAP Logo](https://raw.githubusercontent.com/jordwest/imap-server/master/assets/logo.png)

Go IMAP Server
==============

Barebones IMAP4rev1 server for golang. Designed for integration into a
backend app to provide email client access.

Features a simple API for implementing your own email storage by implementing
golang interfaces. Currently a dummy (in-memory) storage is included, with plans
to include MySQL storage. This would make it simple to integrate into a backend
application to allow users to drag-drop emails into the application, without
messing around with maildir.

Although it would be possible to implement and plug in a maildir storage
interface, that would defeat the purpose of this project and there are much
better, tried and tested open source and commercial solutions that have been
around for a long time (Courier, Dovecot etc).
The goal of this project is to provide simple IMAP access to some kind of existing
system without the overhead of installing a full-blown IMAP/POP3 mail server.


### NOT READY FOR PRODUCTION USE
Currently only plaintext authentication is implemented. This is really bad,
don't use it in any kind of environment where actual passwords or sensitive
emails exists. Actually don't use it anywhere.

Supported Commands
------------------
Command       | Planned  | Implemented  | Tests
------------- | -------  | -----------  | -----
CAPABILITY    | ✓       | ✓           | ✓
NOOP          | ✓       | ✗           | ✗
LOGOUT        | ✓       | ✓           | ✓
AUTHENTICATE  | ✓       | ✓            | ✗
LOGIN         | ✓       | ✓           | ✗
STARTTLS      | ✓       | ✗           | ✗
EXAMINE       | ✓       | ✓           | ✗
CREATE        | ✓       | ✗            | ✗
DELETE        | ✓       | ✗            | ✗
RENAME        | ✓       | ✗            | ✗
SUBSCRIBE     | ✗       | -            | -
UNSUBSCRIBE   | ✗       | -            | -
LIST          | ✓       | ✓           | ✓
LSUB          | ✓       | ✓           | ✓
STATUS        | ✓       | ✓           | ✓
APPEND        | ✓       | ✓           | ✓
CHECK         | ?        | ✗           | ✗
CLOSE         | ✓       | ✓           | ✗
EXPUNGE       | ✓       | ✓           | ✓
SEARCH        | ✓       | ✗           | ✗
FETCH         | ✓       | ✓           | ✓
STORE         | ✓       | ✓           | ✓
COPY          | ✓       | ✗           | ✗
UID           | ✓       | ✓           | ✓
