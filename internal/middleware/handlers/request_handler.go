package handlers

import (
	"github.com/nats-io/nats.go"
)

// Global NATS Variables
var nc *nats.Conn
var js nats.JetStreamContext

func Init(natsConn *nats.Conn, jetStream nats.JetStreamContext) {
	nc = natsConn
	js = jetStream
}
