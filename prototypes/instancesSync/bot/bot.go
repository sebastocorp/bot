package bot

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net"
	"os"
	"prototypes/instancesSync/hashring"
	"slices"
	"sync"
	"time"
)

var ServerInstancesPool = ServerInstancesPoolT{
	servers: map[string]ServerT{},
}

var TransferRequestPool = TransferRequestPoolT{
	pool: map[string]any{},
}

var HashRing = hashring.NewHashRing(1)

type ServerInstancesPoolT struct {
	mu      sync.Mutex
	servers map[string]ServerT
}

type ServerT struct {
	Name    string
	Address string
}

type TransferRequestPoolT struct {
	mu   sync.Mutex
	pool map[string]any
}

type BotT struct {
	Server    ServerT
	ProxyHost string

	Worker WorkerT
	API    APIT
}

// BOT SERVER FUNCTIONS

func (b *BotT) CheckOwnHost() {

	found := false
	for !found {
		log.Printf("check '%s' own host in '%s' proxy host resolution", b.Server.Address, b.ProxyHost)

		time.Sleep(5 * time.Second)

		discoveredHosts, err := net.LookupHost(b.ProxyHost)
		if err != nil {
			log.Printf("unable to look up in '%s' proxy host: %s", b.ProxyHost, err.Error())
			continue
		}

		if slices.Contains(discoveredHosts, b.Server.Address) {
			found = true
			continue
		}

		log.Printf("unable to find '%s' own host in '%s' proxy host resolution", b.Server.Address, b.ProxyHost)
	}
}

func (b *BotT) getOwnIP() (err error) {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		err = fmt.Errorf("error getting network interfaces: %s", err.Error())
		return err
	}

	localAddressList := []string{}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			localAddressList = append(localAddressList, ipNet.IP.To4().String())
		}
	}

	if len(localAddressList) != 1 {
		err = fmt.Errorf("too much IPs from network interfaces '%v'", localAddressList)
		return err
	}

	b.Server.Address = localAddressList[0]

	return err
}

func NewBotServer() (botServer *BotT, err error) {
	botServer = &BotT{
		Server: ServerT{
			Name:    os.ExpandEnv(os.Getenv("BOT_SERVER_NAME")),
			Address: os.ExpandEnv(os.Getenv("BOT_SERVER_ADDRESS")),
		},
		ProxyHost: os.ExpandEnv(os.Getenv("BOT_PROXY_HOST")),
		Worker:    WorkerT{},
		API: APIT{
			Port: os.ExpandEnv(os.Getenv("BOT_SERVER_PORT")),
		},
	}

	if botServer.Server.Name == "" {
		err = fmt.Errorf("server name not provided in 'BOT_SERVER_NAME' environment variable")
		return botServer, err
	}

	if botServer.ProxyHost == "" {
		err = fmt.Errorf("proxy host not provided in 'BOT_PROXY_HOST' environment variable")
		return botServer, err
	}

	if botServer.Server.Address == "" {
		err = botServer.getOwnIP()
		return botServer, err
	}

	if botServer.API.Port == "" {
		botServer.API.Port = "8080"
	}

	return botServer, err
}

func (b *BotT) ShutdownActions(done chan bool, signal chan os.Signal) {
	sig := <-signal
	log.Printf("executing shutdown actions in '%s' with signal '%s'", b.Server.Name, sig.String())

	// stop API
	ctx, cancel := context.WithTimeout(b.API.ctx, 1*time.Second)
	if err := b.API.HttpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	cancel()
	done <- true
}

// SERVER POOL FUNCTIONS

func (sip *ServerInstancesPoolT) GetPool() (result map[string]ServerT) {
	result = map[string]ServerT{}

	sip.mu.Lock()
	maps.Copy(result, sip.servers)
	sip.mu.Unlock()

	return result
}

func (sip *ServerInstancesPoolT) AddServers(servers []ServerT) {
	sip.mu.Lock()
	for _, server := range servers {
		sip.servers[server.Address] = server
	}
	sip.mu.Unlock()
}

func (sip *ServerInstancesPoolT) RemoveServers(servers []ServerT) {
	sip.mu.Lock()
	for _, server := range servers {
		delete(sip.servers, server.Address)
	}
	sip.mu.Unlock()
}

// REQUEST POOL FUNCTIONS

func (trp *TransferRequestPoolT) AddTransferRequest() {

	trp.mu.Lock()
	trp.pool = map[string]any{}
	trp.mu.Unlock()
}

func (trp *TransferRequestPoolT) RemoveTransferRequest(transferKey string) {
	trp.mu.Lock()
	delete(trp.pool, transferKey)
	trp.mu.Unlock()
}

func (trp *TransferRequestPoolT) GetTransferRequestMap() (result map[string]any) {
	result = map[string]any{}

	trp.mu.Lock()
	maps.Copy(result, trp.pool)
	trp.mu.Unlock()

	return result
}
