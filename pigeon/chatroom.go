package pigeon

import (
	"html/template"
	"log"
	"net/http"

	"github.com/ciehanski/pigeon/templates"
)

func (p *Pigeon) chatroom(w http.ResponseWriter, r *http.Request) {
	// Display HTML
	t, err := template.New("chatroom").Parse(templates.ChatroomHTML) // Parse template
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error displaying web page, please try refreshing.", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, nil); err != nil { // Execute template
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error displaying web page, please try refreshing.", http.StatusInternalServerError)
		return
	}

	// Upgrade connection to websocket
	ws, err := p.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgrading connection: %v", err)
		return
	}
	defer func() {
		if err := ws.Close(); err != nil {
			log.Printf("error closing connection: %v", err)
			return
		}
	}()

	// Register the client
	newClient := newClient()
	// Add client to chatroom
	p.Clients[newClient] = ws
	p.Broadcast <- newMessage(newClient, "has connected.")
	log.Printf("client %v has connected\n", newClient.Username)

	// Digest input messages
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error reading message: %v", err)
			p.deleteClient(newClient)
			break
		}
		// Send the newly received message to the broadcast channel
		p.Broadcast <- msg
	}
}

func (p *Pigeon) BroadcastMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-p.Broadcast
		// Send it out to every client that is currently connected
		for client, ws := range p.Clients {
			err := ws.WriteJSON(&msg)
			if err != nil {
				log.Printf("error writing JSON: %v", err)
				if err := ws.Close(); err != nil {
					log.Printf("error closing client: %v", err)
				}
				p.deleteClient(client)
			}
		}
	}
}
