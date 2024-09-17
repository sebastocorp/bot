package bot

import (
	"context"
	"fmt"
	"net"
	"os"
	"slices"
	"time"

	"bot/internal/apiService"
	"bot/internal/database"
	"bot/internal/databaseWorker"
	"bot/internal/global"
	"bot/internal/hashring"
	"bot/internal/hashringWorker"
	"bot/internal/logger"
	"bot/internal/objectStorage"
	"bot/internal/objectWorker"
)

type BotT struct {
	APIService     apiService.APIServiceT
	ObjectWorker   objectWorker.ObjectWorkerT
	DatabaseWorker databaseWorker.DatabaseWorkerT
	HashRingWorker hashringWorker.HashRingWorkerT
}

// BOT SERVER FUNCTIONS

func NewBotServer(config string) (botServer *BotT, err error) {
	botServer = &BotT{}

	err = global.ParseConfig(config)
	if err != nil {
		return botServer, err
	}

	if global.ServerConfig.APIService.Address == "" {
		global.ServerConfig.APIService.Address, err = getOwnIP()
		if err != nil {
			return botServer, err
		}
	}

	if global.ServerConfig.HashRingWorker.Enabled {
		if global.ServerConfig.HashRingWorker.VNodes <= 0 {
			global.ServerConfig.HashRingWorker.VNodes = 1
		}

		global.HashRing = hashring.NewHashRing(global.ServerConfig.HashRingWorker.VNodes)
	}

	if global.ServerConfig.ObjectWorker.ParallelRequests <= 0 {
		global.ServerConfig.ObjectWorker.ParallelRequests = 1
	}

	if global.ServerConfig.DatabaseWorker.ParallelRequests <= 0 {
		global.ServerConfig.DatabaseWorker.ParallelRequests = 1
	}

	if global.ServerConfig.DatabaseWorker.InsertsByConnection <= 0 {
		global.ServerConfig.DatabaseWorker.InsertsByConnection = 1
	}

	ctx := context.Background()
	botServer.ObjectWorker.ObjectManager, err = objectStorage.NewManager(
		ctx,
		objectStorage.S3T{
			Endpoint:        global.ServerConfig.ObjectWorker.ObjectStorage.S3.Endpoint,
			AccessKeyID:     global.ServerConfig.ObjectWorker.ObjectStorage.S3.AccessKeyID,
			SecretAccessKey: global.ServerConfig.ObjectWorker.ObjectStorage.S3.SecretAccessKey,
			Region:          global.ServerConfig.ObjectWorker.ObjectStorage.S3.Region,
			Secure:          global.ServerConfig.ObjectWorker.ObjectStorage.S3.Secure,
		},
		objectStorage.GCST{
			CredentialsFile: global.ServerConfig.ObjectWorker.ObjectStorage.GCS.CredentialsFile,
		},
	)
	if err != nil {
		return botServer, err
	}

	botServer.DatabaseWorker.DatabaseManager, err = database.NewManager(ctx, database.MySQLCredsT{
		Host:     global.ServerConfig.DatabaseWorker.Database.Host,
		Port:     global.ServerConfig.DatabaseWorker.Database.Port,
		User:     global.ServerConfig.DatabaseWorker.Database.Username,
		Pass:     global.ServerConfig.DatabaseWorker.Database.Password,
		Database: global.ServerConfig.DatabaseWorker.Database.Database,
		Table:    global.ServerConfig.DatabaseWorker.Database.Table,
	})

	return botServer, err
}

func (b *BotT) CheckOwnHost() {

	found := false
	for !found {
		time.Sleep(6 * time.Second)

		discoveredHosts, err := net.LookupHost(global.ServerConfig.HashRingWorker.Proxy)
		if err != nil {
			logger.Logger.Errorf("unable to look up in '%s' proxy host: %s", global.ServerConfig.HashRingWorker.Proxy, err.Error())
			continue
		}

		if slices.Contains(discoveredHosts, global.ServerConfig.APIService.Address) {
			found = true
			continue
		}

		logger.Logger.Errorf("unable to find '%s' own host in '%s' proxy host resolution", global.ServerConfig.APIService.Address, global.ServerConfig.HashRingWorker.Proxy)
	}
}

func getOwnIP() (ownAddress string, err error) {

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		err = fmt.Errorf("error getting network interfaces: %s", err.Error())
		return ownAddress, err
	}

	localAddressList := []string{}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			localAddressList = append(localAddressList, ipNet.IP.To4().String())
		}
	}

	if len(localAddressList) != 1 {
		err = fmt.Errorf("too much IPs from network interfaces '%v'", localAddressList)
		return ownAddress, err
	}

	ownAddress = localAddressList[0]

	return ownAddress, err
}

func (b *BotT) ShutdownActions(done chan bool, signal chan os.Signal) {
	sig := <-signal
	logger.Logger.Infof("executing shutdown actions with signal '%s'", sig.String())

	// stop API
	ctx, cancel := context.WithTimeout(b.APIService.Ctx, 1*time.Second)
	if err := b.APIService.HttpServer.Shutdown(ctx); err != nil {
		logger.Logger.Fatalf("error in server Shutdown: %s", err.Error())
	}

	cancel()
	done <- true
}
