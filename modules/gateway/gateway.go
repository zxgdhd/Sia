package gateway

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/NebulousLabs/Sia/build"
	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/persist"
	siasync "github.com/NebulousLabs/Sia/sync"
)

var (
	errNoPeers     = errors.New("no peers")
	errUnreachable = errors.New("peer did not respond to ping")
)

// Gateway implements the modules.Gateway interface.
type Gateway struct {
	listener net.Listener
	myAddr   modules.NetAddress
	port     string

	// handlers are the RPCs that the Gateway can handle.
	handlers map[rpcID]modules.RPCFunc
	// initRPCs are the RPCs that the Gateway calls upon connecting to a peer.
	initRPCs map[string]modules.RPCFunc

	// peers are the nodes we are currently connected to.
	peers map[modules.NetAddress]*peer

	// nodes is the set of all known nodes (i.e. potential peers) on the
	// network.
	nodes map[modules.NetAddress]struct{}

	// threads is used to signal the Gateway's goroutines to shut down and to wait
	// for all goroutines to exit before returning from Close().
	threads siasync.ThreadGroup

	persistDir string

	log *persist.Logger
	mu  sync.RWMutex
}

// Address returns the NetAddress of the Gateway.
func (g *Gateway) Address() modules.NetAddress {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.myAddr
}

// Close saves the state of the Gateway and stops its listener process.
func (g *Gateway) Close() error {
	return g.threads.Stop()
}

// Flush will block until all of the gateway's current processes have finished.
func (g *Gateway) Flush() error {
	return g.threads.Flush()
}

// New returns an initialized Gateway.
func New(addr string, persistDir string) (g *Gateway, err error) {
	// Create the directory if it doesn't exist.
	err = os.MkdirAll(persistDir, 0700)
	if err != nil {
		return
	}

	g = &Gateway{
		handlers:   make(map[rpcID]modules.RPCFunc),
		initRPCs:   make(map[string]modules.RPCFunc),
		peers:      make(map[modules.NetAddress]*peer),
		nodes:      make(map[modules.NetAddress]struct{}),
		persistDir: persistDir,
	}

	// Create the logger.
	g.log, err = persist.NewFileLogger(filepath.Join(g.persistDir, logFile))
	if err != nil {
		return nil, err
	}
	g.threads.AfterStop(func() {
		err = g.log.Close()
		if err != nil {
			// State of the logger is uncertain, a Println will have to
			// suffice.
			fmt.Println("Error when closing the gateway logger:", err)
		}
	})

	// Register RPCs.
	g.RegisterRPC("ShareNodes", g.shareNodes)
	g.RegisterConnectCall("ShareNodes", g.requestNodes)
	g.threads.OnStop(func() {
		g.UnregisterRPC("ShareNodes")
		g.UnregisterConnectCall("ShareNodes")
	})

	// Load the old node list. If it doesn't exist, no problem, but if it does,
	// we want to know about any errors preventing us from loading it.
	if loadErr := g.load(); loadErr != nil && !os.IsNotExist(loadErr) {
		return nil, loadErr
	}
	g.threads.AfterStop(func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		err = g.saveSync()
		if err != nil {
			g.log.Println("Error saving gateway data on shutdown:", err)
		}
	})

	// Add the bootstrap peers to the node list.
	if build.Release == "standard" {
		for _, addr := range modules.BootstrapPeers {
			err := g.addNode(addr)
			if err != nil && err != errNodeExists {
				g.log.Printf("WARN: failed to add the bootstrap node '%v': %v", addr, err)
			}
		}
		g.save()
	}

	// Create listener and set address.
	threadedListenClosedChan := make(chan struct{})
	g.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	// Automatically close the listener when g.threads.Stop() is called.
	g.threads.OnStop(func() {
		err := g.listener.Close()
		if err != nil {
			g.log.Println("WARN: closing the listener failed:", err)
		}
		<-threadedListenClosedChan
	})

	_, g.port, err = net.SplitHostPort(g.listener.Addr().String())
	if err != nil {
		return nil, err
	}
	if build.Release == "testing" {
		g.myAddr = modules.NetAddress(g.listener.Addr().String())
	}

	g.log.Println("INFO: gateway created, started logging")

	// Forward the RPC port, if possible.
	go func() {
		if g.threads.Add() != nil {
			return
		}
		defer g.threads.Done()

		g.managedForwardPort(g.port)
		g.threads.OnStop(func() {
			g.managedClearPort(g.myAddr.Port())
		})
	}()

	// Learn our external IP.
	go g.threadedLearnHostname()

	// Spawn the peer and node managers. These will attempt to keep the peer
	// and node lists healthy.
	go g.threadedPeerManager()
	go g.threadedNodeManager()

	// Spawn the primary listener.
	go g.threadedListen(threadedListenClosedChan)

	return
}

// enforce that Gateway satisfies the modules.Gateway interface
var _ modules.Gateway = (*Gateway)(nil)
