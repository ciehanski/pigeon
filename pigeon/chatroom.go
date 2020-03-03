package pigeon

import (
	"html/template"
	"net/http"
	"time"

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

	// Add the connection to the register channel
	p.Register <- ws

	// Print messages sent prior in this session
	for _, m := range p.BroadcastHistory {
		if err := ws.WriteJSON(m); err != nil {
			p.Log("error writing JSON to websocket: %v", err)
			return
		}
	}

	// Set cookie identifying user
	http.SetCookie(w, &http.Cookie{
		Name:     "clientID",
		Value:    p.Clients[ws].ID,
		Path:     "/",
		Domain:   p.OnionURL,
		Expires:  time.Now().Add(time.Minute * 60),
		Secure:   false,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})

	// Digest messages
	for {
		msg := newMessage(Client{}, "", false)
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			if err.Error() == "websocket: close 1001 (going away)" {
				p.Log("client %v has disconnected\n", msg.Client.Username)
			} else {
				p.Log("error reading message from websocket: %v", err)
			}
			p.Unregister <- ws
			break
		}
		// Add new messages to broadcast history for new users
		p.appendToHistory(msg)
		// Send the newly received message to the broadcast channel
		p.Broadcast <- msg
	}
}
