package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/zenazn/goji"
	"gopkg.in/alecthomas/kingpin.v1"
)

const OK = "250 OK"
const timeout time.Duration = time.Duration(10)
const mail_max_size = 1024 * 1024 * 2 // 2 MB

// In-Memory database
var database []MailConnection

var FORWARD_ENABLED = false

var app = kingpin.New("gosmtpd", "A smtp server for swallowing emails, with possiblity to forward some")

var HOSTNAME = app.Flag("hostname", "Port for the smtp server to listen to").Default("localhost").String()
var PORT = app.Flag("port", "Port for the smtp server to listen to").Default("2525").String()

var forward = app.Command("forwarding", "enable mail forwarding for specific hosts")
var forwardhost = forward.Arg("forwardhost", "The hostname on which email should be forwarded").Required().String()
var forwardport = forward.Arg("forwardport", "The hostname on which email should be forwarded").Required().String()

func handleClient(mc *MailConnection) {
	defer mc.connection.Close()

	greeting := "220 " + *HOSTNAME + " SMTP goSMTPd #" + mc.MailId + " " + time.Now().Format(time.RFC1123Z)
	for i := 0; i < 100; i++ {
		switch mc.state {
		case INITIAL:
			answer(mc, greeting)
			mc.state = NORMAL
		case NORMAL:
			handleNormalMode(mc)
		case READ_DATA:
			var err error
			mc.Data, err = readFrom(mc)
			if err == nil {
				if status := saveMail(mc); status {
					answer(mc, "250 OK : queued as "+mc.MailId)
				} else {
					answer(mc, "554 Error: transaction failed, blame it on the weather")
				}
			} else {
				log.Println("DATA read error: %v", err)
			}
			mc.state = NORMAL
		}

		cont := writeResponse(mc)
		if !cont {
			return
		}
	}
}

func handleNormalMode(mc *MailConnection) {
	line, err := readFrom(mc)
	if err != nil {
		return
	}
	switch {

	case isCommand(line, HELLO):
		if len(line) > 5 {
			mc.helo = line[5:]
		}
		answer(mc, "250 "+*HOSTNAME+" Hello ")

	case isCommand(line, EHLO):
		if len(line) > 5 {
			mc.helo = line[5:]
		}
		answer(mc, "250-"+*HOSTNAME+" Hello "+mc.helo+"["+mc.address+"]"+"\r\n")
		answer(mc, "250-SIZE "+strconv.Itoa(mail_max_size)+"\r\n")
		answer(mc, "250 HELP")

	case isCommand(line, MAIL_FROM):
		if len(line) > 10 {
			mc.From = line[10:]
		}
		answer(mc, OK)

	case isCommand(line, RCPT_TO):
		if len(line) > 8 {
			mc.To = line[8:]
		}
		answer(mc, OK)

	case isCommand(line, DATA):
		answer(mc, "354 Start mail input; end with <CRLF>.<CRLF>")
		mc.state = READ_DATA

	case isCommand(line, NO_OP):
		answer(mc, OK)

	case isCommand(line, RESET):
		mc.From = ""
		mc.To = ""
		answer(mc, OK)

	case isCommand(line, "QUIT"):
		answer(mc, "221 Goodbye")
		killClient(mc)

	default:
		answer(mc, "500 5.5.1 Unrecognized command.")
	}
}

func answer(mc *MailConnection, line string) {
	mc.response = line + "\r\n"
}

func killClient(mc *MailConnection) {
	mc.dropConnection = true
}

func readFrom(mc *MailConnection) (input string, err error) {
	var reply string
	// Command state terminator by default
	suffix := "\r\n"
	if mc.state == READ_DATA {
		// DATA state
		suffix = "\r\n.\r\n"
	}
	for err == nil {
		mc.connection.SetDeadline(time.Now().Add(timeout * time.Second))
		reply, err = mc.reader.ReadString('\n')
		if reply != "" {
			input = input + reply
			if len(input) > mail_max_size {
				err = errors.New("Maximum DATA size exceeded (" + strconv.Itoa(mail_max_size) + ")")
				return input, err
			}
			if mc.state == READ_DATA {
				// Extract the subject while we are at it.
				extractSubject(mc, reply)
			}
		}
		if err != nil {
			break
		}
		if strings.HasSuffix(input, suffix) {
			break
		}
	}
	return input, err
}

// Look at the data block, does it contain a subject
func extractSubject(mc *MailConnection, reply string) {
	if mc.Subject == "" && (len(reply) > 8) {
		test := strings.ToUpper(reply[0:9])
		if strings.Index(test, SUBJECT) == 0 {
			// first line with \r\n
			mc.Subject = reply[9:]
		}
	} else if strings.HasSuffix(mc.Subject, "\r\n") {
		// chop off the \r\n
		mc.Subject = mc.Subject[0 : len(mc.Subject)-2]
		if (strings.HasPrefix(reply, " ")) || (strings.HasPrefix(reply, "\t")) {
			// subject is multi-line
			mc.Subject = mc.Subject + reply[1:]
		}
	}
}

func writeResponse(mc *MailConnection) bool {
	mc.connection.SetDeadline(time.Now().Add(timeout * time.Second))
	size, err := mc.writer.WriteString(mc.response)
	mc.writer.Flush()
	mc.response = mc.response[size:]

	if err != nil {
		if err == io.EOF {
			// connection closed
			return false
		}
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			// a timeout
			return false
		}
	}
	if mc.dropConnection {
		return false
	}
	return true
}

func saveMail(mc *MailConnection) bool {
	for {
		if addr_err := isEmailAddressesValid(mc); addr_err != nil {
			log.Printf("Email from '%s' doesn't have a valid email address the error was %s\n", mc.From, addr_err.Error())
			return false
		}
		database = append(database, *mc)
		if FORWARD_ENABLED {
			if strings.Contains(mc.To, *forwardhost) {
				forwardEmail(mc)
			}
		}
		return true
	}
}

func isEmailAddressesValid(mc *MailConnection) error {
	user, host, addr_err := getEmail(mc.From)
	if addr_err != nil {
		return addr_err
	}
	mc.From = user + "@" + host
	user, host, addr_err = getEmail(mc.To)
	if addr_err != nil {
		return addr_err
	}
	mc.To = user + "@" + host

	return addr_err
}

func getEmail(str string) (name string, host string, err error) {
	address, err := mail.ParseAddress(str)
	if err != nil {
		return "", "", err
	}
	res := strings.Split(address.Address, "@")
	return res[0], res[1], nil
}

func createListener() net.Listener {
	addr := "0.0.0.0:" + *PORT
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Cannot listen on port, %v", err)
	} else {
		log.Println("Listening on tcp ", addr)
	}
	return listener
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case forward.FullCommand():
		FORWARD_ENABLED = true
	}

	setupWebRoutes()


	database = make([]MailConnection, 0)


	listener := createListener()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Panic("Unable to accept: %s", err)
				continue
			}
			go handleClient(&MailConnection{
				connection:    conn,
				address: conn.RemoteAddr().String(),
				reader:   bufio.NewReader(conn),
				writer:  bufio.NewWriter(conn),
				MailId:  uuid.New(),
				Mail:    Mail{Received: time.Now().Unix()},
			})
		}
	}()
	goji.Serve()
}
