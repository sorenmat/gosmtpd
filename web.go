package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

func setupWebRoutes(config *MailConfig) {
	goji.Get("/mail", func(c web.C, w http.ResponseWriter, r *http.Request) { allMails(config, c, w, r) })
	goji.Get("/inbox/:email", func(c web.C, w http.ResponseWriter, r *http.Request) { mails(config, c, w, r) })
	goji.Get("/email/:id", func(c web.C, w http.ResponseWriter, r *http.Request) { mailById(config, c, w, r) })
	goji.Delete("/inbox/:email", func(c web.C, w http.ResponseWriter, r *http.Request) { deleteMails(config, c, w, r) })
	goji.Delete("/email/:id", func(c web.C, w http.ResponseWriter, r *http.Request) { deleteById(config, c, w, r) })

}
func allMails(config *MailConfig, c web.C, w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	encoder.Encode(config.database)
}

func mails(config *MailConfig, c web.C, w http.ResponseWriter, r *http.Request) {
	email := c.URLParams["email"]
	encoder := json.NewEncoder(w)

	result := make([]MailConnection, 0)
	for _, msg := range config.database {
		if msg.To == email {
			result = append(result, msg)
		}
	}
	encoder.Encode(result)
}

func mailById(config *MailConfig, c web.C, w http.ResponseWriter, r *http.Request) {
	id := c.URLParams["id"]
	encoder := json.NewEncoder(w)
	for _, msg := range config.database {
		if msg.MailId == id {
			encoder.Encode(msg)
		}
	}
}

func deleteMails(config *MailConfig, c web.C, w http.ResponseWriter, r *http.Request) {
	email := c.URLParams["email"]

	result := make([]MailConnection, 0)
	for _, msg := range config.database {
		if msg.To != email {
			result = append(result, msg)
		}
	}
	config.database = result
}

func deleteById(config *MailConfig, c web.C, w http.ResponseWriter, r *http.Request) {
	id := c.URLParams["id"]

	result := make([]MailConnection, 0)
	log.Println("Deleting ", id)
	for _, msg := range config.database {
		if msg.MailId != id {
			result = append(result, msg)
		}
	}
	config.database = result
}
