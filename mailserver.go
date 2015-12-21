package main

import (
	"code.google.com/p/go-uuid/uuid"
	"log"
	"strings"
	"sync"
	"time"
)

type MailServer struct {
	hostname       string
	port           string
	forwardEnabled bool
	forwardHost    string
	forwardPort    string
	// In-Memory database
	database []MailConnection
	mu       *sync.Mutex

	httpport       string
	expireinterval int
}

func (server *MailServer) saveMail(mail *MailConnection) bool {
	server.mu.Lock()
	defer server.mu.Unlock()
	if err := isEmailAddressesValid(mail); err != nil {
		log.Printf("Email from '%s' doesn't have a valid email address the error was %s\n", mail.From, err.Error())
		return false
	}
	for _, rcpt := range mail.recepient {
		to, err := cleanupEmail(rcpt)
		if err != nil {
			log.Println("Cleaning up email gave an error ", err)
		}
		mail.To = to
		mail.expireStamp = time.Now().Add(time.Duration(mail.mailserver.expireinterval) * time.Second)
		mail.Received = time.Now().Unix()
		mail.MailId = uuid.New()
		server.database = append(server.database, *mail)

	}
	if *forwardhost != "" && *forwardport != "" {
		if strings.Contains(mail.To, *forwardhost) {
			forwardEmail(mail)
		}
	}

	return true
}

func (server *MailServer) cleanupDatabase() {
	server.mu.Lock()
	dbcopy := []MailConnection{}
	log.Printf("About to expire entries from database older then %d seconds, current length %d\n", server.expireinterval, len(server.database))

	for _, v := range server.database {
		if time.Since(v.expireStamp).Seconds() < 0 {
			dbcopy = append(dbcopy, v)
		}
	}
	server.database = dbcopy
	log.Println("After expire entries in database, new length ", len(server.database))
	server.mu.Unlock()
}
