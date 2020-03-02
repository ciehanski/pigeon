package pigeon

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string    `json:"id"`
	Client    *Client   `json:"client"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func newMessage(client *Client, message string) Message {
	return Message{
		ID:        uuid.New().String(),
		Client:    client,
		Message:   message,
		Timestamp: time.Now(),
	}
}
