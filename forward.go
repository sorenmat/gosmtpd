package main

import (
	"net"
	"net/smtp"
)

func getForwardHost() string {
	return net.JoinHostPort(*forwardhost, *forwardport)
}
func forwardEmail(client *MailConnection) {
	smtp.SendMail(getForwardHost(), nil, client.From, []string{client.To}, []byte(client.Data))

}
