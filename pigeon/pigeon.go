package pigeon

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/gorilla/websocket"
	"github.com/ipsn/go-libtor"
)

type Pigeon struct {
	Clients     map[*Client]*websocket.Conn
	Broadcast   chan Message
	Server      *http.Server
	Upgrader    websocket.Upgrader
	OnionURL    string
	RemotePort  int
	TorVersion3 bool
	Debug       bool
}

func (p *Pigeon) Init(ctx context.Context) (*tor.Tor, *tor.OnionService, error) {
	// Make pigeon instance
	p.Clients = make(map[*Client]*websocket.Conn)
	p.Upgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Start Tor
	t, err := startTor()
	if err != nil {
		return nil, nil, err
	}

	// Start listening over onion service
	onionSvc, err := p.listenTor(ctx, t)
	if err != nil {
		return nil, nil, err
	}

	// Init serving
	http.HandleFunc("/", p.chatroom)
	p.Server = &http.Server{
		// Tor is quite slow and depending on the size of the files being
		// transferred, the server could timeout. I would like to keep set timeouts, but
		// will need to find a sweet spot or enable an option for large transfers.
		IdleTimeout:  time.Minute * 3,
		ReadTimeout:  time.Minute * 3,
		WriteTimeout: time.Minute * 3,
		Handler:      nil,
	}

	return t, onionSvc, nil
}

func startTor() (*tor.Tor, error) {
	var tempDataDir string
	if runtime.GOOS != "windows" {
		tempDataDir = "/tmp"
	} else {
		tempDataDir = "%TEMP%"
	}

	t, err := tor.Start(nil, &tor.StartConf{ // Start tor
		ProcessCreator: libtor.Creator,
		//DebugWriter:            os.Stderr,
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
	// Create an onion service to listen on any port but show as 80
	onionSvc, err := t.Listen(ctx, &tor.ListenConf{
		Version3:    p.TorVersion3,
		RemotePorts: []int{p.RemotePort},
	})
	if err != nil {
		return nil, err
	}
	return onionSvc, nil
}

func (p *Pigeon) deleteClient(client *Client) {
	p.Broadcast <- newMessage(client, "has disconnected.")
	delete(p.Clients, client)
}
