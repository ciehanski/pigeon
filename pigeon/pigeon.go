package pigeon

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/gorilla/websocket"
	"github.com/ipsn/go-libtor"
)

// Pigeon is the object that stores all clients, messages, and
// configuration options.
type Pigeon struct {
	Clients          map[*websocket.Conn]Client
	Broadcast        chan Message
	BroadcastHistory []Message
	Register         chan *websocket.Conn
	Unregister       chan *websocket.Conn
	Server           *http.Server
	Upgrader         *websocket.Upgrader
	OnionURL         string
	RemotePort       int
	TorVersion3      bool
	Logger           *log.Logger
	Debug            bool
}

// Init initializes the pigeon object and all of its options and dependencies.
func (p *Pigeon) Init(ctx context.Context) (*tor.Tor, *tor.OnionService, error) {
	// Make pigeon instance
	p.Clients = make(map[*websocket.Conn]Client)
	p.Register = make(chan *websocket.Conn, 10)
	p.Unregister = make(chan *websocket.Conn, 10)
	p.Broadcast = make(chan Message, 10)
	p.Upgrader = &websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		HandshakeTimeout:  time.Second * 12,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	if p.Debug {
		p.Logger = log.New(os.Stdout, "[pigeon-debug] ", log.LstdFlags)
	} else {
		p.Logger = log.New(os.Stderr, "[pigeon] ", log.LstdFlags)
	}

	// Start Tor
	t, err := p.startTor()
	if err != nil {
		return nil, nil, err
	}

	// Start listening over onion service
	onionSvc, err := p.listenTor(ctx, t)
	if err != nil {
		return nil, nil, err
	}

	// Init serving
	rtr := http.NewServeMux()
	rtr.HandleFunc("/", p.chatroom)
	rtr.HandleFunc("/ws", p.websocket)
	p.Server = &http.Server{
		// Tor is quite slow and depending on the size of the files being
		// transferred, the server could timeout. I would like to keep set timeouts, but
		// will need to find a sweet spot or enable an option for large transfers.
		IdleTimeout:  time.Minute * 3,
		ReadTimeout:  time.Minute * 3,
		WriteTimeout: time.Minute * 3,
		Handler:      rtr,
	}
	return t, onionSvc, nil
}

// BroadcastMessages continuously loops and checks if it needs to add
// a new client, remove a current client, or broadcast any message received
//from a client to all other clients.
func (p *Pigeon) BroadcastMessages() {
	for {
		select {
		case conn := <-p.Register:
			// Register the client
			newClient := newClient()
			// Add client to chatroom
			p.Clients[conn] = newClient
			// Broadcast that a new user has connected
			connMsg := newMessage(newClient, "has connected.", true)
			// Toss connMsg into the Broadcast channel to be sent to all other clients
			p.Broadcast <- connMsg
			// Add message to broadcast history
			p.BroadcastHistory = append(p.BroadcastHistory, connMsg)
			// Log it!
			p.Log("client %v has connected\n", newClient.Username)
		case conn := <-p.Unregister:
			dconnMsg := newMessage(p.Clients[conn], "has disconnected.", false)
			// Toss dconnMsg into the Broadcast channel to be sent to all other clients
			p.Broadcast <- dconnMsg
			// Add message to broadcast history
			p.BroadcastHistory = append(p.BroadcastHistory, dconnMsg)
			// Log it!
			p.Log("client %v has disconnected\n", p.Clients[conn].Username)
			// Remove client from clients map
			delete(p.Clients, conn)
			// Close the socket
			if err := conn.Close(); err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					p.Log("error closing websocket: %v", err)
				}
			}
		case msg := <-p.Broadcast:
			// Send it out to every client that is currently connected
			for ws := range p.Clients {
				err := ws.WriteJSON(msg)
				if err != nil {
					p.Log("error writing JSON to websocket: %v", err)
					p.Unregister <- ws
					if err := ws.Close(); err != nil {
						p.Log("error closing websocket: %v", err)
					}
				}
			}
		}
	}
}

// Log is a helper function to only log if debug is enabled.
func (p *Pigeon) Log(str string, args ...interface{}) {
	if p.Debug {
		p.Logger.Printf(str, args...)
	}
}

func (p *Pigeon) startTor() (*tor.Tor, error) {
	var tempDataDir string
	if runtime.GOOS != "windows" {
		tempDataDir = "/tmp/pigeon"
	} else {
		tempDataDir = `%TEMP%\pigeon`
	}

	t, err := tor.Start(nil, &tor.StartConf{ // Start tor
		ProcessCreator:         libtor.Creator,
		DebugWriter:            p.Logger.Writer(),
		UseEmbeddedControlConn: runtime.GOOS != "windows", // This option is not supported on Windows
		TempDataDirBase:        tempDataDir,
		DataDir:                tempDataDir,
		RetainTempDataDir:      false,
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (p *Pigeon) listenTor(ctx context.Context, t *tor.Tor) (*tor.OnionService, error) {
	onionSvc, err := t.Listen(ctx, &tor.ListenConf{
		Version3:    p.TorVersion3,
		RemotePorts: []int{p.RemotePort},
	})
	if err != nil {
		return nil, err
	}
	return onionSvc, nil
}
