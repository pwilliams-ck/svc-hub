package main

import (
	"context"
	"log"
	"time"

	"github.com/cloudkey-io/service-hub/logger-svc/data"
)

// Methods that take this as a receiver are available over RPC, as long as they
// are exported.
type RPCServer struct{}

type RPCPayload struct {
	Name string
	Data string
}

// LogInfo logs an entry to the database.
func (r *RPCServer) LogInfo(payload RPCPayload, res *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error inserting log entry: ", err)
		return err
	}

	*res = "processed payload via RPC: " + payload.Name
	return nil
}
