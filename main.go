package main

import (
	"bufio"
	"log"
	"net"

	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/zenazn/goji"
	"gopkg.in/alecthomas/kingpin.v2"
)

const timeout time.Duration = time.Duration(10)
const mailMaxSize = 1024 * 1024 * 2 // 2 MB

var webport = kingpin.Flag("webport", "Port the web server should run on").Default("8000").String()
var HOSTNAME = kingpin.Flag("hostname", "Hostname for the smtp server to listen to").Default("localhost").String()
var PORT = kingpin.Flag("port", "Port for the smtp server to listen to").Default("2525").String()

var forwardhost = kingpin.Flag("forwardhost", "The hostname on which email should be forwarded").Default("").OverrideDefaultFromEnvar("FORWARD_HOST").String()
var forwardport = kingpin.Flag("forwardport", "The port on which email should be forwarded").Default("25").OverrideDefaultFromEnvar("FORWARD_PORT").String()

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
	log.Println("Trying to bind web to port ",*webport)
	l, lerr := net.Listen("tcp", ":"+*webport)
	if lerr != nil {
		log.Fatalf("Unable to bind to port %s.\n%s\n. ",*webport, lerr)
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
	
	serve(MailConfig{hostname: *HOSTNAME, port: *PORT, forwardEnabled: forwardEnabled, forwardHost: *forwardhost, forwardPort: *forwardport})
}
