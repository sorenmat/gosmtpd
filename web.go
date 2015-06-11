package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

func setupWebRoutes() {
	goji.Get("/mail", allMails)
	goji.Get("/inbox/:email", mails)
	goji.Get("/email/:id", mailById)
	goji.Delete("/inbox/:email", deleteMails)
	goji.Delete("/email/:id", deleteById)

}
func allMails(c web.C, w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	encoder.Encode(database)
}

func mails(c web.C, w http.ResponseWriter, r *http.Request) {
	email := c.URLParams["email"]
	encoder := json.NewEncoder(w)

	result := make([]MailConnection, 0)
	for _, msg := range database {
		if msg.To == email {
			result = append(result, msg)
		}
	}
	encoder.Encode(result)
}

func mailById(c web.C, w http.ResponseWriter, r *http.Request) {
	id := c.URLParams["id"]
	encoder := json.NewEncoder(w)
	result := make([]MailConnection, 0)
	for _, msg := range database {
		if msg.MailId == id {
			result = append(result, msg)
		}
	}
	encoder.Encode(result)
}

func deleteMails(c web.C, w http.ResponseWriter, r *http.Request) {
	email := c.URLParams["email"]

	result := make([]MailConnection, 0)
	for _, msg := range database {
		if msg.To != email {
			result = append(result, msg)
		}
	}
	database = result
}

func deleteById(c web.C, w http.ResponseWriter, r *http.Request) {
	id := c.URLParams["id"]

	result := make([]MailConnection, 0)
	fmt.Println("Trying to delete ", id)
	for _, msg := range database {
		if msg.MailId != id {
			result = append(result, msg)
		}
	}
	database = result
}
