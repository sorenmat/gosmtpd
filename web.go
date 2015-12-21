package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

func setupWebRoutes(config *MailServer) {
	goji.Get("/status", func(c web.C, w http.ResponseWriter, r *http.Request) { status(config, c, w, r) })
	goji.Get("/mail", func(c web.C, w http.ResponseWriter, r *http.Request) { allMails(config, c, w, r) })
	goji.Get("/inbox/:email", func(c web.C, w http.ResponseWriter, r *http.Request) { inbox(config, c, w, r) })
	goji.Get("/email/:id", func(c web.C, w http.ResponseWriter, r *http.Request) { mailByID(config, c, w, r) })
	goji.Delete("/inbox/:email", func(c web.C, w http.ResponseWriter, r *http.Request) { deleteMails(config, c, w, r) })
	goji.Delete("/email/:id", func(c web.C, w http.ResponseWriter, r *http.Request) { deleteByID(config, c, w, r) })

}

func status(config *MailServer, c web.C, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func allMails(config *MailServer, c web.C, w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	encoder.Encode(config.database)
}

func inbox(config *MailServer, c web.C, w http.ResponseWriter, r *http.Request) {
	email := c.URLParams["email"]
	encoder := json.NewEncoder(w)

	var result []MailConnection
	for _, msg := range config.database {
		if msg.To == email {
			result = append(result, msg)
		}
	}
	if len(result) == 0 {
		http.NotFound(w, r)
	}

	encoder.Encode(result)
}

func mailByID(config *MailServer, c web.C, w http.ResponseWriter, r *http.Request) {
	id := c.URLParams["id"]
	encoder := json.NewEncoder(w)
	found := false
	for _, msg := range config.database {
		if msg.MailId == id {
			encoder.Encode(msg)
			found = true
		}
	}
	if !found {
		http.NotFound(w, r)
	}
}

func deleteMails(config *MailServer, c web.C, w http.ResponseWriter, r *http.Request) {
	email := c.URLParams["email"]

	var result []MailConnection
	for _, msg := range config.database {
		if msg.To != email {
			result = append(result, msg)
		}
	}
	config.database = result
}

func deleteByID(config *MailServer, c web.C, w http.ResponseWriter, r *http.Request) {
	id := c.URLParams["id"]

	var result []MailConnection
	log.Println("Deleting ", id)
	for _, msg := range config.database {
		if msg.MailId != id {
			result = append(result, msg)
		}
	}
	config.database = result
}
