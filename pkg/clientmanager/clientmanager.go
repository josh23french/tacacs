package clientmanager

import (
    "log"
    "net"
    "runtime"
    "strings"
    "sync"
    "time"
	"github.com/josh23french/tacacs/pkg/config"
)

type ClientManager struct {
    sync.RWMutex
    updaterTicker  *time.Ticker
    closeUpdater chan bool
    clients []*Client
    lookup map[string]*Client
}

type Client struct {
    sync.RWMutex
    Hostname string
    Addrs    []string
    Secret   string
}

func New(clientsFromConfig []config.Client) *ClientManager {
    clients := make([]*Client, 0)
    for _, client := range clientsFromConfig {
        clients = append(clients, &Client{
            Hostname: client.Hostname,
            Secret: client.Secret,
        })
    }
    
    cm := &ClientManager{
        updaterTicker: time.NewTicker(60 * time.Second),
        clients: clients,
        closeUpdater: make(chan bool),
        lookup: make(map[string]*Client, 0),
    }
    cm.update()
    go func() {
        for {
            select {
            case <-cm.closeUpdater:
                return
            case <-cm.updaterTicker.C:
                cm.update()
            }
        }
    }()
    runtime.SetFinalizer(cm, (*ClientManager).Close)
    return cm
}

func (cm *ClientManager) update() {
    log.Printf("Starting ClientManager update...\n")
    cm.RLock()
    for _, c := range cm.clients {
        c.RLock()
        hostname := c.Hostname
        c.RUnlock()

        addrs, err := net.LookupHost(hostname)
        if err != nil {
            log.Printf("Error looking up %v: %v\n", c.Hostname, err)
            continue // Leave this host as-is... Or do we try again??
        }

        c.Lock()
        c.Addrs = addrs
        c.Unlock()
        
        cm.RUnlock()
        cm.Lock()
        for _, addr := range addrs {
            cm.lookup[addr] = c
        }
        cm.Unlock()
        cm.RLock()
    }
    log.Printf("ClientManager update finished: %+v\n", cm.lookup)
    cm.RUnlock()
}

func (cm *ClientManager) Close() {
    cm.updaterTicker.Stop()
    cm.closeUpdater <- true
}

func (cm *ClientManager) FindByAddr(addr net.Addr) *Client {
    addrParts := strings.SplitN(addr.String(), ":", 2)
    if len(addrParts) != 2 {
        log.Printf("Could not parse client addr: %v", addr.String())
        return nil
    }
    clientIP := net.ParseIP(addrParts[0])
    if clientIP == nil {
        log.Printf("Could not parse client IP: %v", addrParts[0])
        return nil
    }
    
    cm.RLock()
    client := cm.lookup[clientIP.String()]
    cm.RUnlock()
    return client
}