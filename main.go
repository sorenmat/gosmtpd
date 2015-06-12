package main

import (
	"bufio"
	"errors"
	"fmt"
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

func (mc *MailConnection) resetDeadLine() {
	mc.connection.SetDeadline(time.Now().Add(timeout * time.Second))
}

func (mc *MailConnection) nextLine() (string, error) {
	return mc.reader.ReadString('\n')
}

func (mc *MailConnection) inDataMode() bool {
	return mc.state == READ_DATA
}

// Flush the response buffer to the n
func (mc *MailConnection) flush() bool {
	mc.resetDeadLine()
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

func processClientRequest(mc *MailConnection) {
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
				log.Printf("DATA read error: %v\n", err)
			}
			mc.state = NORMAL
		}

		cont := mc.flush()
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

func getTerminateString(state State) string {
	terminateString := "\r\n"
	if state == READ_DATA {
		// DATA state
		terminateString = "\r\n.\r\n"
	}
	return terminateString

}

func readFrom(mc *MailConnection) (input string, err error) {
	var read string

	terminateString := getTerminateString(mc.state)

	for err == nil {
		mc.resetDeadLine()
		read, err = mc.nextLine()
		if err != nil {
			break
		}

		if read != "" {
			input = input + read
			if len(input) > mail_max_size {
				err = errors.New("DATA size exceeded (" + strconv.Itoa(mail_max_size) + ")")
				return input, err
			}

			if mc.inDataMode() {
				scanForSubject(mc, read)
			}
		}
		if strings.HasSuffix(input, terminateString) {
			break
		}
	}
	return input, err
}

func scanForSubject(mc *MailConnection, line string) {
	if mc.Subject == "" && strings.Index(strings.ToUpper(line), SUBJECT) == 0 {
		mc.Subject = line[9:]
	} else if strings.HasSuffix(mc.Subject, "\r\n") {
		// remove the newline stuff
		mc.Subject = mc.Subject[0 : len(mc.Subject)-2]
		if (strings.HasPrefix(line, " ")) || (strings.HasPrefix(line, "\t")) {
			// multiline subject
			mc.Subject = mc.Subject + line[1:]
		}
	}
}

func saveMail(mc *MailConnection) bool {
	log.Println("Saving email")
	if err := isEmailAddressesValid(mc); err != nil {
		log.Printf("Email from '%s' doesn't have a valid email address the error was %s\n", mc.From, err.Error())
		return false
	}
	database = append(database, *mc)
	if FORWARD_ENABLED {
		if strings.Contains(mc.To, *forwardhost) {
			forwardEmail(mc)
		}
	}
	fmt.Println("Email saved !")
	return true
}

func isEmailAddressesValid(mc *MailConnection) error {
	fmt.Println("Calling is email valid")
	if from, err := cleanupEmail(mc.From); err == nil {
		mc.From = from
	} else {
		return err
	}
	if to, err := cleanupEmail(mc.To); err == nil {
		mc.To = to
	} else {
		return err
	}
	return nil
}

func cleanupEmail(str string) (email string, err error) {
	address, err := mail.ParseAddress(str)
	if err != nil {
		return "", err
	}
	return address.Address, nil
}

func createListener() net.Listener {
	addr := "0.0.0.0:" + *PORT
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Cannot listen on port\n, %v", err)
	} else {
		log.Println("Listening on tcp ", addr)
	}
	return listener
}

func serve() {

	setupWebRoutes()

	database = make([]MailConnection, 0)

	listener := createListener()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Panicf("Unable to accept: %s\n", err)
				continue
			}
			go processClientRequest(&MailConnection{
				connection: conn,
				address:    conn.RemoteAddr().String(),
				reader:     bufio.NewReader(conn),
				writer:     bufio.NewWriter(conn),
				MailId:     uuid.New(),
				Mail:       Mail{Received: time.Now().Unix()},
			})
		}
	}()
	goji.Serve()
}
func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case forward.FullCommand():
		FORWARD_ENABLED = true
	}
	serve()
}
