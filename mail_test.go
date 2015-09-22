package main

import (
	"log"
	"net/smtp"
	"testing"
)

func TestCleanUpEmail(t *testing.T) {
	email := "<sorenm@mymessages.dk>"
	ce, err := cleanupEmail(email)
	if err != nil {
		t.Error(err)
	}
	if ce != "sorenm@mymessages.dk" {
		t.Error("Address is wrong, when cleaning emails")
	}
}

func init() {
	/*	log.Println("MailTest init")
			os.Setenv("WEBPORT", "8383")
			os.Setenv("FORWARD_HOST", "smtp.gmail.com")
			os.Setenv("FORWARD_SMtP", "tradeshift.com")
			kingpin.Parse()

			go serve(MailConfig{port: "2626", httpport: "8383", forwardEnabled: true})
		SendMailOnPort("smo@tradeshift.com", "2525")
	*/
}
func SendMailOnPort(receiver string, port string) {
	err := smtp.SendMail("localhost:"+port,
		nil,
		"sorenm@mymessages.dk", // sender
		[]string{receiver},     //recipient
		[]byte("Subject: Testing\nThis is $the email body.\nAnd it is the bomb"),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func TestIsEmailAddressesValid(t *testing.T) {

}
