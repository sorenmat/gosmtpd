package main

import (
	"bufio"
	"log"
	"net"
	"sync"

	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/zenazn/goji"
	"gopkg.in/alecthomas/kingpin.v2"
)

const timeout time.Duration = time.Duration(10)
const mailMaxSize = 1024 * 1024 * 2 // 2 MB

var webport = kingpin.Flag("webport", "Port the web server should run on").Default("8000").OverrideDefaultFromEnvar("WEBPORT").String()
var hostname = kingpin.Flag("hostname", "Hostname for the smtp server to listen to").Default("localhost").String()
var port = kingpin.Flag("port", "Port for the smtp server to listen to").Default("2525").String()

var forwardhost = kingpin.Flag("forwardhost", "The hostname after the @ that we should forward i.e. gmail.com").Default("").OverrideDefaultFromEnvar("FORWARD_HOST").String()
var forwardsmtp = kingpin.Flag("forwardsmtp", "SMTP server to forward the mail to").Default("").OverrideDefaultFromEnvar("FORWARD_SMTP").String()
var forwardport = kingpin.Flag("forwardport", "The port on which email should be forwarded").Default("25").OverrideDefaultFromEnvar("FORWARD_PORT").String()
var forwarduser = kingpin.Flag("forwarduser", "The username for the forward host").Default("").OverrideDefaultFromEnvar("FORWARD_USER").String()
var forwardpassword = kingpin.Flag("forwardpassword", "Password for the user").Default("").OverrideDefaultFromEnvar("FORWARD_PASSWORD").String()

var cleanupInterval = kingpin.Flag("mailexpiration", "Time in seconds for a mail to expire, and be removed from database").Default("300").Int()

func createListener(server *MailServer) net.Listener {
	addr := "0.0.0.0:" + server.port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	} else {
		log.Println("Listening on tcp ", addr)
	}
	return listener
}

func serve(mailserver *MailServer) {
	if mailserver.httpport == "" {
		log.Fatal("HTTPPort needs to be configured")
	}
	mailserver.database = make([]MailConnection, 0)
	mailserver.mu = &sync.Mutex{}
	go func() {
		for {
			mailserver.cleanupDatabase()
			time.Sleep(time.Duration(mailserver.expireinterval) * time.Second)
		}
	}()
	setupWebRoutes(mailserver)

	listener := createListener(mailserver)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Panicf("Unable to accept: %s\n", err)
				continue
			}
			go processClientRequest(&MailConnection{
				mailserver: mailserver,
				connection: conn,
				address:    conn.RemoteAddr().String(),
				reader:     bufio.NewReader(conn),
				writer:     bufio.NewWriter(conn),
				MailId:     uuid.New(),
				Mail:       Mail{Received: time.Now().Unix()},
			})
		}
	}()
	log.Println("Trying to bind web to port ", mailserver.httpport)
	l, lerr := net.Listen("tcp", ":"+mailserver.httpport)
	if lerr != nil {
		log.Fatalf("Unable to bind to port %s.\n%s\n. ", mailserver.httpport, lerr)
	}
	goji.ServeListener(l)
}

func main() {
	kingpin.Version("0.1.1")
	kingpin.Parse()
	forwardEnabled := false
	if *forwardhost != "" && *forwardport != "" {
		forwardEnabled = true
	}
	log.Println("Clenaup interval ", *cleanupInterval)

	serve(&MailServer{hostname: *hostname, port: *port, httpport: *webport, forwardEnabled: forwardEnabled, forwardHost: *forwardhost, forwardPort: *forwardport, expireinterval: *cleanupInterval})
}
