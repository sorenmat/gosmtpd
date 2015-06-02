package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"testing"
	"time"
)

func getPort() *string {
	test := "2525"
	return &test
}
func TestSendingMail(t *testing.T) {
	PORT = getPort() // we need to force this, since we don't parse the commandline
	go serve()
	time.Sleep(2 * time.Second)
	SendMail("soren@test.com")
	resp, _ := http.Get("http://localhost:8000/inbox/soren@test.com")
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

func getEmailByHash(hash string, t *testing.T) {
	resp, _ := http.Get("http://localhost:8000/email/" + hash)
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
		name, host, err := getEmail(email)
		if name != "test" || host != "something.com" {
			t.Error()
		}
		if err != nil {
			t.Error()
		}
	}
}
