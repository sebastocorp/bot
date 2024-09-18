package bot

import (
	"context"
	"os"
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
		global.Config.APIService.Address = "0.0.0.0"
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

func (b *BotT) InitServer() {
	// Init bot server
	b.HashRingWorker.InitWorker()
	b.ObjectWorker.InitWorker()
	b.DatabaseWorker.InitWorker()
	b.APIService.InitAPI()

	for !global.ServerState.IsReady() {
		logger.Logger.Infof("waiting for bot server ready...")
		time.Sleep(5 * time.Second)
	}

	logger.Logger.Infof("bot server is ready")
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
