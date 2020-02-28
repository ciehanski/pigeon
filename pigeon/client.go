package pigeon

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
)

type Client struct {
	ID       string
	Username string
}

func newClient() *Client {
	return &Client{
		ID:       uuid.New().String(),
		Username: randomdata.SillyName(),
	}
}
