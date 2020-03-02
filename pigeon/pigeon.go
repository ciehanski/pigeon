package pigeon

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/gorilla/websocket"
	"github.com/ipsn/go-libtor"
)

type Pigeon struct {
	Clients          map[*websocket.Conn]*Client
	Broadcast        chan Message
	BroadcastHistory []Message
	Server           *http.Server
	Upgrader         websocket.Upgrader
	OnionURL         string
	RemotePort       int
	TorVersion3      bool
	Logger           *log.Logger
	Debug            bool
}

func (p *Pigeon) Init(ctx context.Context) (*tor.Tor, *tor.OnionService, error) {
	// Make pigeon instance
	p.Clients = make(map[*websocket.Conn]*Client)
	p.Upgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		HandshakeTimeout:  time.Second * 10,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	if p.Debug {
		p.Logger = log.New(os.Stdout, "[pigeon] ", log.LstdFlags)
	} else {
		p.Logger = log.New(os.Stderr, "[pigeon] ", log.LstdFlags)
	}
	p.Broadcast = make(chan Message, 10)

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

func (p *Pigeon) BroadcastMessages() {
	for {
		select {
		case msg := <-p.Broadcast:
			// Send it out to every client that is currently connected
			for ws := range p.Clients {
				err := ws.WriteJSON(msg)
				if err != nil {
					p.Log("error writing JSON: %v", err)
					p.deleteClient(ws)
					if err := ws.Close(); err != nil {
						p.Log("error closing websocket: %v", err)
					}
				}
			}
		}
	}
}

func (p *Pigeon) Log(str string, args ...interface{}) {
	if p.Debug {
		p.Logger.Printf(str, args...)
	}
}

func (p *Pigeon) startTor() (*tor.Tor, error) {
	var tempDataDir string
	if runtime.GOOS != "windows" {
		tempDataDir = "/tmp"
	} else {
		tempDataDir = "%TEMP%"
	}

	t, err := tor.Start(nil, &tor.StartConf{ // Start tor
		ProcessCreator:         libtor.Creator,
		DebugWriter:            p.Logger.Writer(),
		UseEmbeddedControlConn: runtime.GOOS != "windows", // This option is not supported on Windows
		TempDataDirBase:        tempDataDir,
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

func (p *Pigeon) deleteClient(conn *websocket.Conn) {
	p.Broadcast <- newMessage(p.Clients[conn], "has disconnected.")
	p.BroadcastHistory = append(p.BroadcastHistory, newMessage(p.Clients[conn], "has disconnected."))
	delete(p.Clients, conn)
}
