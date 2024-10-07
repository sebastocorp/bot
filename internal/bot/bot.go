package bot

import (
	"context"
	"os"
	"time"

	"bot/api/v1alpha2"
	"bot/internal/components/apiService"
	"bot/internal/components/databaseWorker"
	"bot/internal/components/hashringWorker"
	"bot/internal/components/objectWorker"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/pools"
)

type BotT struct {
	config v1alpha2.BOTConfigT
	log    logger.LoggerT

	APIService     *apiService.APIServiceT
	ObjectWorker   *objectWorker.ObjectWorkerT
	DatabaseWorker *databaseWorker.DatabaseWorkerT
	HashringWorker *hashringWorker.HashringWorkerT
}

// BOT SERVER FUNCTIONS

func NewBotServer(configFilepath string) (botServer *BotT, err error) {
	botConfig, err := parseConfig(configFilepath)
	if err != nil {
		return botServer, err
	}

	botServer = &BotT{
		config: botConfig,
	}

	err = botServer.checkConfig()
	if err != nil {
		return botServer, err
	}

	logCommon := global.GetLogCommonFields()
	logCommon[global.LogFieldKeyCommonInstance] = botServer.config.Name
	botServer.log = logger.NewLogger(context.Background(),
		logger.GetLevel(botServer.config.LogLevel),
		logCommon,
	)

	dbPool := pools.NewDatabaseRequestPool()
	objectPool := pools.NewObjectRequestPool()
	serverPool := pools.NewServerPool()

	botServer.APIService = apiService.NewApiService(&botServer.config, objectPool)

	botServer.ObjectWorker, err = objectWorker.NewObjectWorker(&botServer.config, objectPool, dbPool)
	if err != nil {
		return botServer, err
	}

	botServer.DatabaseWorker, err = databaseWorker.NewDatabaseWorker(&botServer.config, dbPool)
	if err != nil {
		return botServer, err
	}

	botServer.HashringWorker = hashringWorker.NewHashringWorker(&botServer.config, serverPool)

	return botServer, err
}

func (b *BotT) Run() {
	// Init bot server
	b.HashringWorker.Run()
	b.ObjectWorker.Run()
	b.DatabaseWorker.Run()
	b.APIService.Run()

	for !global.ServerState.IsReady() {
		b.log.Debug("waiting for bot server ready...", map[string]any{})
		time.Sleep(5 * time.Second)
	}

	b.log.Info("bot server is ready", map[string]any{})
}

func (b *BotT) ShutdownActions(done chan bool, signal chan os.Signal) {
	sig := <-signal

	b.log.Debug("executing shutdown actions", map[string]any{
		"signal": sig.String(),
	})

	b.APIService.Shutdown()
	b.DatabaseWorker.Shutdown()
	b.ObjectWorker.Shutdown()
	b.HashringWorker.Shutdown()

	done <- true
}
