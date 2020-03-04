package pigeon

import (
	"time"

	"github.com/google/uuid"
)

// Message contains all fields required to send a message
// with pigeon.
type Message struct {
	ID        string    `json:"id"`
	Client    Client    `json:"client"`
	Message   string    `json:"message"`
	Connected bool      `json:"connected"`
	Timestamp time.Time `json:"timestamp"`
}

func newMessage(client Client, message string, connected bool) Message {
	return Message{
		ID:        uuid.New().String(),
		Client:    client,
		Message:   message,
		Connected: connected,
		Timestamp: time.Now(),
	}
}
