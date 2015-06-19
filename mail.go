package main

import (
	"errors"
	"io"
	"log"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

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

	greeting := "220 " + mc.mailconfig.hostname + " SMTP goSMTPd #" + mc.MailId + " " + time.Now().Format(time.RFC1123Z)
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
					answer(mc, "554 Error: transaction failed.")
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
	log.Println("Line: ", line)
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
		answer(mc, "250-SIZE "+strconv.Itoa(mailMaxSize)+"\r\n")
		answer(mc, "250 HELP")

	case isCommand(line, MAIL_FROM):
		if len(line) > 10 {
			mc.From = line[10:]
		}
		answer(mc, OK)

	case isCommand(line, RCPT_TO):
		if len(line) > 8 {
			mc.recepient = append(mc.recepient, line[8:])
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
			if len(input) > mailMaxSize {
				err = errors.New("DATA size exceeded (" + strconv.Itoa(mailMaxSize) + ")")
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
	for _, rcpt := range mc.recepient {
		to, err := cleanupEmail(rcpt)
		if err != nil {
			log.Println("Cleaning up email gave an error ", err)
		}
		mc.To = to
		mc.mailconfig.database = append(mc.mailconfig.database, *mc)

	}
	if ForwardEnabled {
		if strings.Contains(mc.To, *forwardhost) {
			forwardEmail(mc)
		}
	}
	log.Println("Email saved !")
	return true
}

func isEmailAddressesValid(mc *MailConnection) error {
	if from, err := cleanupEmail(mc.From); err == nil {
		mc.From = from
	} else {
		log.Println("From is not valid")
		return err
	}
	if to, err := cleanupEmail(mc.To); err == nil {
		mc.To = to
	} else {
		log.Println("To is not valid: ", mc.To)
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
