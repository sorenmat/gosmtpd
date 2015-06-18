package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/zenazn/goji"
	"gopkg.in/alecthomas/kingpin.v1"
)

const timeout time.Duration = time.Duration(10)
const mailMaxSize = 1024 * 1024 * 2 // 2 MB

var ForwardEnabled = false

var app = kingpin.New("gosmtpd", "A smtp server for swallowing emails, with possiblity to forward some")

var HOSTNAME = app.Flag("hostname", "Port for the smtp server to listen to").Default("localhost").String()
var PORT = app.Flag("port", "Port for the smtp server to listen to").Default("2525").String()

var forward = app.Command("forwarding", "enable mail forwarding for specific hosts")
var forwardhost = forward.Arg("forwardhost", "The hostname on which email should be forwarded").Required().String()
var forwardport = forward.Arg("forwardport", "The hostname on which email should be forwarded").Required().String()

func createListener(config MailConfig) net.Listener {
	addr := "0.0.0.0:" + config.port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		//log.Fatalf("Cannot listen on '%s' port '%s'\nerror: %v", addr, config.port, err)
		panic(err)
	} else {
		log.Println("Listening on tcp ", addr)
	}
	return listener
}

func serve(config MailConfig) {
	config.database = make([]MailConnection, 0)

	setupWebRoutes(&config)

	listener := createListener(config)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Panicf("Unable to accept: %s\n", err)
				continue
			}
			go processClientRequest(&MailConnection{
				mailconfig: &config,
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
		ForwardEnabled = true
	}
	serve(MailConfig{hostname: *HOSTNAME, port: *PORT, forwardEnabled: ForwardEnabled, forwardHost: *forwardhost, forwardPort: *forwardport})
}
