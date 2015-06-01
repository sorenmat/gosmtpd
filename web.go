package main
import (
	"encoding/json"
	"net/http"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji"
)

func setupWebRoutes() {
	goji.Get("/mail", allMails)
	goji.Get("/inbox/:email", mails)
	goji.Get("/email/:id", mailById)

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

