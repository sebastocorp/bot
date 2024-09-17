package bot

import (
	"context"
	"fmt"
	"net"
	"os"
	"slices"
	"time"

	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/database"
	"bot/internal/managers/objectStorage"
	"bot/internal/workers/apiService"
	"bot/internal/workers/databaseWorker"
	"bot/internal/workers/hashringWorker"
	"bot/internal/workers/objectWorker"
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

	if global.Config.APIService.Address == "" {
		global.Config.APIService.Address, err = getOwnIP()
		if err != nil {
			return botServer, err
		}
	}

	if global.Config.HashRingWorker.Enabled {
		if global.Config.HashRingWorker.VNodes <= 0 {
			global.Config.HashRingWorker.VNodes = 1
		}
	}

	if global.Config.ObjectWorker.ParallelRequests <= 0 {
		global.Config.ObjectWorker.ParallelRequests = 1
	}

	if global.Config.DatabaseWorker.ParallelRequests <= 0 {
		global.Config.DatabaseWorker.ParallelRequests = 1
	}

	if global.Config.DatabaseWorker.InsertsByConnection <= 0 {
		global.Config.DatabaseWorker.InsertsByConnection = 1
	}

	ctx := context.Background()
	botServer.ObjectWorker.ObjectManager, err = objectStorage.NewManager(
		ctx,
		global.Config.ObjectWorker.ObjectStorage.S3,
		global.Config.ObjectWorker.ObjectStorage.GCS,
	)
	if err != nil {
		return botServer, err
	}

	botServer.DatabaseWorker.DatabaseManager, err = database.NewManager(ctx,
		global.Config.DatabaseWorker.Database,
	)

	return botServer, err
}

func (b *BotT) CheckOwnHost() {

	found := false
	for !found {
		time.Sleep(6 * time.Second)

		discoveredHosts, err := net.LookupHost(global.Config.HashRingWorker.Proxy)
		if err != nil {
			logger.Logger.Errorf("unable to look up in '%s' proxy host: %s", global.Config.HashRingWorker.Proxy, err.Error())
			continue
		}

		if slices.Contains(discoveredHosts, global.Config.APIService.Address) {
			found = true
			continue
		}

		logger.Logger.Errorf("unable to find '%s' own host in '%s' proxy host resolution", global.Config.APIService.Address, global.Config.HashRingWorker.Proxy)
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
