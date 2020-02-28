package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ciehanski/pigeon/pigeon"
)

func main() {
	var p pigeon.Pigeon

	// Init flags
	flag.BoolVar(&p.Debug, "debug", false, "run in debug mode")
	flag.BoolVar(&p.TorVersion3, "torv3", true, "use version 3 of the Tor circuit (recommended)")
	flag.IntVar(&p.RemotePort, "port", 80, "remote port used to host the onion service")
	flag.Parse()

	// Wait at most 3 minutes to publish the service
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Init Tor connection
	t, onionSvc, err := p.Init(ctx)
	if err != nil {
		log.Fatalf("Error starting Tor & initializing onion service: %v", err)
	}
	defer func() {
		if err = onionSvc.Close(); err != nil {
			log.Fatalf("Error closing connection to onion service: %v", err)
		}
		if err = t.Close(); err != nil {
			log.Fatalf("Error closing connection to Tor: %v", err)
		}
	}()

	// Display the onion service URL
	p.OnionURL = onionSvc.ID
	fmt.Printf("Please open a Tor capable browser and navigate to http://%v.onion\n", p.OnionURL)

	// Start listening for incoming chat messages and broadcast them
	go p.BroadcastMessages()

	srvErrCh := make(chan error, 1)
	go func() { srvErrCh <- p.Server.Serve(onionSvc) }() // Begin serving
	if err = <-srvErrCh; err != nil {
		log.Fatalf("Error serving on onion service: %v", err)
	}
	defer func() { // Proper server shutdown when program ends
		if err = p.Server.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down pigeon server: %v", err)
		}
	}()
}
