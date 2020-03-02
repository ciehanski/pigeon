package pigeon

import (
	"html/template"
	"net/http"

	"github.com/ciehanski/pigeon/templates"
)

func (p *Pigeon) chatroom(w http.ResponseWriter, r *http.Request) {
	// Display HTML
	t, err := template.New("chatroom").Parse(templates.ChatroomHTML) // Parse template
	if err != nil {
		p.Log("Error parsing template: %v", err)
		http.Error(w, "Error displaying web page, please try refreshing.", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, nil); err != nil { // Execute template
		p.Log("Error executing template: %v", err)
		http.Error(w, "Error displaying web page, please try refreshing.", http.StatusInternalServerError)
		return
	}
}

func (p *Pigeon) websocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to websocket
	ws, err := p.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.Log("error upgrading connection: %v", err)
		return
	}
	defer func() {
		if err := ws.Close(); err != nil {
			p.Log("error closing connection: %v", err)
			return
		}
	}()

	// Print messages sent prior in this session
	for _, m := range p.BroadcastHistory {
		if err := ws.WriteJSON(m); err != nil {
			p.Log("error writing to websocket: %v", err)
			return
		}
	}
	// Register the client
	newClient := newClient()
	// TODO: data race?
	// Add client to chatroom
	p.Clients[ws] = newClient
	// Broadcast that a new user has connected
	p.Broadcast <- newMessage(newClient, "has connected.")
	p.BroadcastHistory = append(p.BroadcastHistory, newMessage(newClient, "has connected."))
	p.Log("client %v has connected\n", newClient.Username)

	// Digest messages
	for {
		msg := newMessage(nil, "")
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			p.Log("error reading message from websocket: %v", err)
			p.deleteClient(ws)
			break
		}
		// Add new messages to broadcast history for new users
		p.BroadcastHistory = append(p.BroadcastHistory, msg)
		// Send the newly received message to the broadcast channel
		p.Broadcast <- msg
	}
}
