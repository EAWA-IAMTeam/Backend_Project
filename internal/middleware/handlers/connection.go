package handlers

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

// Global NATS & JetStream variables
var nc *nats.Conn
var js nats.JetStreamContext

// Initialize NATS and JetStream
func InitNATS() error {
	var err error
	nc, err = nats.Connect("nats://192.168.0.189:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
		return err
	}

	// Create JetStream Context
	js, err = nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to initialize JetStream: %v", err)
		return err
	}

	fmt.Println("NATS & JetStream successfully initialized")
	return nil
}

// GetNATSConnection returns the NATS connection
func GetNATSConnection() *nats.Conn {
	return nc
}

// GetJetStreamContext returns the JetStream context
func GetJetStreamContext() nats.JetStreamContext {
	return js
}
