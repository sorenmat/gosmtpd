package main

import (
	"log"
	"net"
	"net/smtp"
)

func getForwardHost() string {
	return net.JoinHostPort(*forwardsmtp, *forwardport)
}
func forwardEmail(client *MailConnection) {
	log.Printf("Forwarding mail to: %v using host %v ", client.To, getForwardHost())

	err := smtp.SendMail(getForwardHost(), getAuth(), client.From, []string{client.To}, []byte(client.Data))
	if err != nil {
		log.Println("Forward error: ", err)
	}
}

func getAuth() smtp.Auth {
	if *forwarduser == "" && *forwardpassword == "" {
		return nil
	}
	auth := smtp.PlainAuth(
		"",
		*forwarduser,
		*forwardpassword,
		*forwardsmtp,
	)
	return auth

}
