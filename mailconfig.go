package main

import "sync"

type MailConfig struct {
	hostname       string
	port           string
	forwardEnabled bool
	forwardHost    string
	forwardPort    string
	// In-Memory database
	database MailDatabase
	mu       *sync.Mutex
}
