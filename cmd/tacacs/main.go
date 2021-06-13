package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/josh23french/tacplus"
	"github.com/josh23french/tacacs/pkg/config"
	"github.com/josh23french/tacacs/pkg/clientmanager"
	"github.com/josh23french/tacacs/pkg/backends"
)

func serveConnection(config *config.Config) func(net.Conn) {
    clients := clientmanager.New(config.Clients)
    backends := backends.NewManager(config.Backends)
    handler := NewHandler(config, backends)
    
    return func(nc net.Conn) {
        client := clients.FindByAddr(nc.RemoteAddr())
        
        if client == nil { // Unknown client
            nc.Close() // Who dis?!
            return
        }
    	s := tacplus.ServerConnHandler{
    		Handler: handler,
    		ConnConfig: tacplus.ConnConfig{
    			Secret: []byte(client.Secret),
    			Mux:    true,
    		},
    	}
    	s.Serve(nc)
    }
}

func main() {
	var (
		configFile = flag.String("config.file", "/etc/tacacs/config.yaml", "path to config file")
	)
	flag.Parse()

	config := config.NewConfig(*configFile)
	
	listen := config.Server.Listen
	
	if listen == "" {
	    listen = "0.0.0.0:49"
	}

	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("Couldn't listen: %v", fmt.Errorf("%w", err))
	}
	
	log.Printf("Listening on %v", listen)
	srv := &tacplus.Server{
		ServeConn: serveConnection(config),
	}
	srv.Serve(l)
}
