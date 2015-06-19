package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"testing"
	"time"
)

func init() {
	os.Setenv("PORT", "8282")

	go serve(MailConfig{port: "2525", forwardEnabled: false})
	time.Sleep(2 * time.Second)

}

func getPort() *string {
	test := "2525"
	return &test
}

func getMailConnection(email string) []MailConnection {
	resp, httperr := http.Get("http://localhost:8000/inbox/" + email)
	if httperr != nil {
		panic("Unable to call rest")
	}
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	decoder := json.NewDecoder(resp.Body)
	var d []MailConnection
	err := decoder.Decode(&d)
	if err != nil {
		panic("Unable to decode message list")
	}
	return d
}

func BenchmarkSendMails(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SendMailWithMessage("sorenbench@test.com", "message")

	}
	d := getMailConnection("sorenbench@test.com")
	if b.N != len(d) {
		b.Errorf("Wrong number of email expected %d got %d\n", b.N, len(d))
	}

}

func TestSendingMailWithMultilines(t *testing.T) {
	//	PORT = getPort() // we need to force this, since we don't parse the commandline
	SendMailWithMessage("sorenz@test.com", `This
is
a
test`)
	time.Sleep(1 * time.Second)
	d := getMailConnection("sorenz@test.com")

	if len(d) != 1 {
		t.Error("To many email ", len(d))
	}

	email := getEmailByHash(d[0].MailId, t)
	if email.To == "" {
		t.Error("To should not be empty")
	}
}

func TestSendingMail(t *testing.T) {

	SendMail("sorenm@test.com")
	d := getMailConnection("sorenm@test.com")

	if len(d) != 1 {
		t.Error("To many email")
	}
	getEmailByHash(d[0].MailId, t)
}

func TestSendingMailAndDeletingIt(t *testing.T) {

	SendMail("sorenm1@test.com")
	d := getMailConnection("sorenm1@test.com")

	if len(d) != 1 {
		t.Error("Expected one email got ", len(d))
	}

	deleteEmailById(d[0].MailId)
	d = getMailConnection("sorenm1@test.com")

	if len(d) != 0 {
		t.Errorf("Not the correct number '%d' of emails\n ", len(d))
	}

}

func getEmailByHash(hash string, t *testing.T) MailConnection {
	resp, _ := http.Get("http://localhost:8000/email/" + hash)
	if resp.StatusCode != 200 {
		t.Error(resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	var d MailConnection
	err := decoder.Decode(&d)
	if err != nil {
		t.Error("Unable to decode message list")
	}

	return d
}

func deleteRequest(url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(
		"DELETE",
		url,
		bytes.NewBuffer([]byte("")),
	)
	req.Header.Set("Content-Type", "application/json")
	reply, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	return reply, err
}

// Delete a specific mail by finding via the id
func deleteEmailById(hash string) {
	deleteRequest("http://localhost:8000/email/" + hash)
}

// Delete all mails in an inbox
func emptyMailBox(email string) {
	deleteRequest("http://localhost:8000/inbox/" + email)
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
		t.Error("Expected something.com:2525 got ", result)
	}
}

func TestForwardHostnameWithoutPort(t *testing.T) {
	host := "something.com"
	port := ""
	result := net.JoinHostPort(host, port)
	if result != "something.com:" {
		t.Error("Expected 'something.com:' got ", result)
	}
}

func TestEmail(t *testing.T) {
	emails := []string{"test@something.com", "  test@something.com", " test@something.com        "}
	for _, email := range emails {
		name, err := cleanupEmail(email)
		if name != "test@something.com" {
			t.Error("Expected 'test@something.com' got ", name)
		}
		if err != nil {
			t.Error("Cleanup email resulted in an error ", err)
		}
	}
}

func TestExtractSubjectOnSingleLine(t *testing.T) {
	mc := &MailConnection{}
	line := `SUBJECT: Testing`
	scanForSubject(mc, line)
	if mc.Subject != "Testing" {
		t.Error("Expected 'Testing' got ", mc.Subject)
	}
}
func TestExtractSubjectOnSingleNotFormattedLine(t *testing.T) {
	mc := &MailConnection{}
	line := `SUBJECT:             Testing                  `
	scanForSubject(mc, line)
	if mc.Subject != "            Testing                  " {
		t.Error("Exptected '            Testing                  ' got ", mc.Subject)
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
