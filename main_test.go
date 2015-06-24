package main

import (
	"bytes"
	"crypto/rand"
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
	//time.Sleep(1 * time.Second)

}

func getPort() *string {
	test := "2525"
	return &test
}

func getMailConnection(email string) ([]MailConnection, int) {
	resp, _ := http.Get("http://localhost:8000/inbox/" + email)

	decoder := json.NewDecoder(resp.Body)
	var d []MailConnection
	decoder.Decode(&d)
	return d, resp.StatusCode
}

func BenchmarkSendMails(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SendMailWithMessage("sorenbench@test.com", "message")

	}
	d, _ := getMailConnection("sorenbench@test.com")
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
	d, _ := getMailConnection("sorenz@test.com")

	if len(d) != 1 {
		t.Error("To many email ", len(d))
	}

	email, _ := getEmailByHash(d[0].MailId)
	if email.To == "" {
		t.Error("To should not be empty")
	}
}

func TestSendingMail(t *testing.T) {

	SendMail("sorenm@test.com")
	d, _ := getMailConnection("sorenm@test.com")

	if len(d) != 1 {
		t.Error("To many email")
	}
	getEmailByHash(d[0].MailId)
}

func isMailBoxSize(email string, size int, t *testing.T) bool {
	d, _ := getMailConnection(email)
	if len(d) != size {
		t.Errorf("Wrong number of emails, expected %d but was %d", size, len(d))
		return false
	}
	return true
}

func TestSendingMailsToMultipleReceivers(t *testing.T) {

	SendMails([]string{"meet@test.com", "joe@test.com", "black@test.com"})

	isMailBoxSize("meet@test.com", 1, t)

	isMailBoxSize("joe@test.com", 1, t)
	isMailBoxSize("black@test.com", 1, t)

}

func TestSendingMailsToMultipleReceiversAndDeletingThem(t *testing.T) {

	SendMails([]string{"meet1@test.com", "joe1@test.com", "black1@test.com"})
	isMailBoxSize("meet1@test.com", 1, t)
	emptyMailBox("meet1@test.com")
	isMailBoxSize("meet1@test.com", 0, t)

	isMailBoxSize("joe1@test.com", 1, t)
	emptyMailBox("joe1@test.com")
	isMailBoxSize("joe1@test.com", 0, t)

	isMailBoxSize("black1@test.com", 1, t)
	emptyMailBox("black1@test.com")
	isMailBoxSize("black1@test.com", 0, t)
}

func TestSendingMailsToMultipleReceiversAndDeletingById(t *testing.T) {

	SendMails([]string{"meet2@test.com", "joe2@test.com", "black1@test.com"})

	mail, _ := getMailConnection("meet2@test.com")
	if len(mail) != 1 {
		t.Error("Wrong number of emails")
	}
	deleteEmailByID(mail[0].MailId)
	mails, _ := getMailConnection("meet2@test.com")
	if len(mails) != 0 {
		t.Error("Should be empty ,but was ", len(mails))
	}
	mails, _ = getMailConnection("joe2@test.com")
	if len(mails) == 0 {
		t.Error("Was empty, but shouldn't be")
	}

}

func TestSendingMailAndDeletingIt(t *testing.T) {
	email := randomEmail()
	SendMail(email)
	d, _ := getMailConnection(email)

	if len(d) != 1 {
		t.Error("Expected one email got ", len(d))
	}

	deleteEmailByID(d[0].MailId)
	d, _ = getMailConnection(email)

	if len(d) != 0 {
		t.Errorf("Not the correct number '%d' of emails\n ", len(d))
	}

}

func TestGettingANonExistingInbox(t *testing.T) {
	email := randomEmail()
	_, status := getMailConnection(email)
	if status != 404 {
		t.Error("Should return 404")
	}

}

func TestGettingANonExistingId(t *testing.T) {
	email := randomEmail()
	_, status := getEmailByHash(email)
	if status != 404 {
		t.Error("Should return 404, but was ", status)
	}

}

func getEmailByHash(hash string) (MailConnection, int) {
	resp, _ := http.Get("http://localhost:8000/email/" + hash)

	decoder := json.NewDecoder(resp.Body)
	var d MailConnection
	decoder.Decode(&d)

	return d, resp.StatusCode
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
func deleteEmailByID(hash string) {
	deleteRequest("http://localhost:8000/email/" + hash)
}

// Delete all mails in an inbox
func emptyMailBox(email string) {
	deleteRequest("http://localhost:8000/inbox/" + email)
}

func SendMails(receiver []string) {
	err := smtp.SendMail("localhost:2525",
		nil,
		"sorenm@mymessages.dk", // sender
		receiver,               //recipient
		[]byte("Subject: Testing\nThis is $the email body.\nAnd it is the bomb"),
	)
	if err != nil {
		log.Fatal(err)
	}
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

func randomEmail() string {
	var dictionary string

	//	if randType == "alphanum" {
	dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	//	}

	/*        if randType == "alpha" {
	        dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
	        dictionary = "0123456789"
	}
	*/
	var bytes = make([]byte, 10)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes) + "@test.com"
}
