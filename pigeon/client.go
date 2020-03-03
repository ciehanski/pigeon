package pigeon

import (
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
)

// Client contains all fields required for pigeon.
type Client struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	ConnectedAt time.Time `json:"connected_at"`
}

func newClient() Client {
	return Client{
		ID:          uuid.New().String(),
		Username:    randomdata.SillyName(),
		ConnectedAt: time.Now(),
	}
}
