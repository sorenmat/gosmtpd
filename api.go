package main
import (
	"net"
	"bufio"
)


type Mail struct {
	To, From, Subject, Data string
	Received                int64
}

type MailConnection struct {
	Mail
	state          State
	helo           string
	response       string
	address        string
	MailId         string

	connection     net.Conn
	reader         *bufio.Reader
	writer         *bufio.Writer
	dropConnection bool

}

type State int

const (
	INITIAL   State = iota
	NORMAL    State = iota
	READ_DATA State = iota
)

const (
	EHLO = "EHLO"
	NO_OP = "NOOP"
	HELLO = "HELO"
	SUBJECT = "SUBJECT: "
	DATA = "DATA"
	MAIL_FROM = "MAIL FROM:"
	RCPT_TO = "RCPT TO:"
	RESET = "RSET"

)
