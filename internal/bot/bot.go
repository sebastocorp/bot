package bot

import (
	"context"
	"fmt"
	"maps"
	"net"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"

	"bot/internal/database"
	"bot/internal/hashring"
	"bot/internal/logger"
	"bot/internal/objectStorage"
)

var ServerInstancesPool = ServerInstancesPoolT{
	servers: map[string]ServerT{},
}

var TransferRequestPool = TransferRequestPoolT{
	pool: map[string]TransferT{},
}

var HashRing = hashring.NewHashRing(1)

type BotT struct {
	Server           ServerT
	ProxyHost        string
	ObjectManager    objectStorage.ManagerT
	DatabaseManager  database.ManagerT
	ParallelRequests int

	API APIT
}

type ServerInstancesPoolT struct {
	mu      sync.Mutex
	servers map[string]ServerT
}

type ServerT struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type TransferRequestPoolT struct {
	mu   sync.Mutex
	pool map[string]TransferT
}

type TransferT struct {
	From objectStorage.ObjectT `json:"from"`
	To   objectStorage.ObjectT `json:"to"`
}

// BOT SERVER FUNCTIONS

func (b *BotT) CheckOwnHost() {

	found := false
	for !found {
		time.Sleep(6 * time.Second)

		discoveredHosts, err := net.LookupHost(b.ProxyHost)
		if err != nil {
			logger.Logger.Errorf("unable to look up in '%s' proxy host: %s", b.ProxyHost, err.Error())
			continue
		}

		if slices.Contains(discoveredHosts, b.Server.Address) {
			found = true
			continue
		}

		logger.Logger.Errorf("unable to find '%s' own host in '%s' proxy host resolution", b.Server.Address, b.ProxyHost)
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

func (b *BotT) ShutdownActions(done chan bool, signal chan os.Signal) {
	sig := <-signal
	logger.Logger.Infof("executing shutdown actions in '%s' with signal '%s'", b.Server.Name, sig.String())

	// stop API
	ctx, cancel := context.WithTimeout(b.API.ctx, 1*time.Second)
	if err := b.API.HttpServer.Shutdown(ctx); err != nil {
		logger.Logger.Fatalf("error in server Shutdown: %s", err.Error())
	}

	cancel()
	done <- true
}

func NewBotServer() (botServer *BotT, err error) {
	botServer = &BotT{
		Server: ServerT{
			Name:    os.ExpandEnv(os.Getenv("BOT_SERVER_NAME")),
			Address: os.ExpandEnv(os.Getenv("BOT_SERVER_ADDRESS")),
		},
		ProxyHost: os.ExpandEnv(os.Getenv("BOT_SERVER_PROXY_HOST")),
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

	botServer.ParallelRequests = 1
	parallelRequest := os.Getenv("BOT_WORKER_PARALLEL_REQUESTS")
	if parallelRequest != "" {
		botServer.ParallelRequests, err = strconv.Atoi(parallelRequest)
		if err != nil {
			err = fmt.Errorf("invalid environment variable 'BOT_WORKER_PARALLEL_REQUESTS' value: %s", err.Error())
			return botServer, err
		}
	}

	if botServer.ParallelRequests <= 0 {
		err = fmt.Errorf("invalid environment variable 'BOT_WORKER_PARALLEL_REQUESTS' value, must be a positive number")
		return botServer, err
	}

	// ParallelRequestNumber

	ctx := context.Background()
	botServer.ObjectManager, err = objectStorage.NewManager(
		ctx,
		objectStorage.S3T{
			Endpoint:        os.ExpandEnv(os.Getenv("BOT_SERVER_S3_ENDPOINT")),
			AccessKeyID:     os.ExpandEnv(os.Getenv("BOT_SERVER_S3_ACCESS_KEY")),
			SecretAccessKey: os.ExpandEnv(os.Getenv("BOT_SERVER_S3_SECRET_ACCESS_KEY")),
		},
		objectStorage.GCST{
			CredentialsFile: os.ExpandEnv(os.Getenv("BOT_SERVER_GCS_CREDENTIALS_FILE")),
		},
	)
	if err != nil {
		return botServer, err
	}

	botServer.DatabaseManager, err = database.NewManager(ctx, database.MySQLCredsT{
		Host:     os.ExpandEnv(os.Getenv("BOT_SERVER_DB_HOST")),
		Port:     os.ExpandEnv(os.Getenv("BOT_SERVER_DB_PORT")),
		User:     os.ExpandEnv(os.Getenv("BOT_SERVER_DB_USER")),
		Pass:     os.ExpandEnv(os.Getenv("BOT_SERVER_DB_PASS")),
		Database: os.ExpandEnv(os.Getenv("BOT_SERVER_DB_DATABASE")),
		Table:    os.ExpandEnv(os.Getenv("BOT_SERVER_DB_TABLE")),
	})

	return botServer, err
}

// SERVER POOL FUNCTIONS

func (sip *ServerInstancesPoolT) GetPool() (result map[string]ServerT) {
	result = map[string]ServerT{}

	sip.mu.Lock()
	maps.Copy(result, sip.servers)
	sip.mu.Unlock()

	return result
}

func (sip *ServerInstancesPoolT) GetServersList() (result []ServerT) {
	pool := sip.GetPool()
	result = []ServerT{}

	for _, server := range pool {
		result = append(result, server)
	}

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

func (trp *TransferRequestPoolT) AddTransferRequest(transfer TransferT) {
	trp.mu.Lock()
	trp.pool[transfer.To.ObjectPath] = transfer
	trp.mu.Unlock()
}

func (trp *TransferRequestPoolT) RemoveTransferRequest(transferKey string) {
	trp.mu.Lock()
	delete(trp.pool, transferKey)
	trp.mu.Unlock()
}

func (trp *TransferRequestPoolT) GetTransferRequestMap() (result map[string]TransferT) {
	result = map[string]TransferT{}

	trp.mu.Lock()
	maps.Copy(result, trp.pool)
	trp.mu.Unlock()

	return result
}
