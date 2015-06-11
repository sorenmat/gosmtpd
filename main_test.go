package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/smtp"
	"testing"
	"time"
)

func init() {
	PORT = getPort() // we need to force this, since we don't parse the commandline

	go serve()
	time.Sleep(2 * time.Second)

}

func getPort() *string {
	test := "2525"
	return &test
}

func TestSendingMailWithMultilines(t *testing.T) {
	//	PORT = getPort() // we need to force this, since we don't parse the commandline
	SendMailWithMessage("sorenz@test.com", `This
is
a
test`)
	time.Sleep(1 * time.Second)
	resp, _ := http.Get("http://localhost:8000/inbox/sorenz@test.com")
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
	data, _ := httputil.DumpResponse(resp, true)
	fmt.Println(string(data))
	decoder := json.NewDecoder(resp.Body)
	var d []MailConnection
	err := decoder.Decode(&d)
	if err != nil {
		t.Error("Unable to decode message list")
	}

	if len(d) != 1 {
		t.Error("To many email ", len(d))
	}

	email := getEmailByHash(d[0].MailId, t)
	fmt.Println(email[0].Data)
}

func TestSendingMail(t *testing.T) {

	SendMail("sorenm@test.com")
	resp, _ := http.Get("http://localhost:8000/inbox/sorenm@test.com")
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
	decoder := json.NewDecoder(resp.Body)
	var d []MailConnection
	err := decoder.Decode(&d)
	if err != nil {
		t.Error("Unable to decode message list")
	}

	if len(d) != 1 {
		t.Error("To many email")
	}

	getEmailByHash(d[0].MailId, t)
}

func getEmailByHash(hash string, t *testing.T) []MailConnection {
	resp, _ := http.Get("http://localhost:8000/email/" + hash)
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}
	dump, _ := httputil.DumpResponse(resp, true)
	fmt.Println(string(dump))
	decoder := json.NewDecoder(resp.Body)
	var d []MailConnection
	err := decoder.Decode(&d)
	if err != nil {
		t.Error("Unable to decode message list")
	}

	if len(d) != 1 {
		t.Error("To many email")
	}
	return d
}

func SendMail(receiver string) {
	err := smtp.SendMail("localhost:2525",
		nil,
		"sorenm@mymessages.dk", // sender
		[]string{receiver},     //recipient
		[]byte("Subject: Testing\nThis is $the email body.\nAnd it is the bomb"),
	)
	if err != nil {
		log.Fatal(err)
	}
}
func SendMailWithMessage(receiver string, msg string) {
	err := smtp.SendMail("localhost:2525",
		nil,
		"sorenm@mymessages.dk", // sender
		[]string{receiver},     //recipient
		[]byte(msg),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func TestForwardHostname(t *testing.T) {
	host := "something.com"
	port := "2525"
	result := net.JoinHostPort(host, port)
	if result != "something.com:2525" {
		t.Error()
	}
}

func TestForwardHostnameWithoutPort(t *testing.T) {
	host := "something.com"
	port := ""
	result := net.JoinHostPort(host, port)
	if result != "something.com:" {
		t.Error()
	}
}

func TestEmail(t *testing.T) {
	emails := []string{"test@something.com", "  test@something.com", " test@something.com        "}
	for _, email := range emails {
		name, err := cleanupEmail(email)
		if name != "test@something.com" {
			t.Error()
		}
		if err != nil {
			t.Error()
		}
	}
}

func TestExtractSubjectOnSingleLine(t *testing.T) {
	mc := &MailConnection{}
	line := `SUBJECT: Testing`
	scanForSubject(mc, line)
	if mc.Subject != "Testing" {
		t.Error()
	}
}
func TestExtractSubjectOnSingleNotFormattedLine(t *testing.T) {
	mc := &MailConnection{}
	line := `SUBJECT:             Testing                  `
	scanForSubject(mc, line)
	if mc.Subject != "            Testing                  " {
		t.Error()
	}
}

func TestExtractSubjectWithNewLine(t *testing.T) {
	mc := &MailConnection{}
	line := `SUBJECT: 
Testing`
	scanForSubject(mc, line)
	if mc.Subject != "\nTesting" {
		t.Error(mc.Subject, len(mc.Subject))
	}
}

func TestExtractSubjectOnMultipleLines(t *testing.T) {
	mc := &MailConnection{}
	line := `SUBJECT: 
Testing
Multiline subject`
	scanForSubject(mc, line)
	if mc.Subject != "\nTesting\nMultiline subject" {
		t.Error(mc.Subject, len(mc.Subject))
	}
}
