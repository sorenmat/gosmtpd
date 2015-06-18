package main

import (
	"bufio"
	"net"
)

type MailConfig struct {
	hostname       string
	port           string
	forwardEnabled bool
	forwardHost    string
	forwardPort    string
	// In-Memory database
	database []MailConnection
}

// Basic structure of the mail, this is used for serializing the mail to the storage
type Mail struct {
	To, From, Subject, Data string
	Received                int64
}

// When a connection is made to the server, a MailConnection object is made, to keep track
// of the specific client connection
type MailConnection struct {
	Mail
	state    State
	helo     string
	response string
	address  string
	MailId   string

	connection     net.Conn
	reader         *bufio.Reader
	writer         *bufio.Writer
	dropConnection bool
	mailconfig     *MailConfig
}

// State representating the flow of the connection
type State int

const (
	INITIAL   State = iota
	NORMAL    State = iota
	READ_DATA State = iota
)

const (
	OK        = "250 OK"
	EHLO      = "EHLO"
	NO_OP     = "NOOP"
	HELLO     = "HELO"
	SUBJECT   = "SUBJECT: "
	DATA      = "DATA"
	MAIL_FROM = "MAIL FROM:"
	RCPT_TO   = "RCPT TO:"
	RESET     = "RSET"
)
