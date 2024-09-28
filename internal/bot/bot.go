package bot

import (
	"context"
	"log"
	"os"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/components/apiService"
	"bot/internal/components/databaseWorker"
	"bot/internal/components/hashringWorker"
	"bot/internal/components/objectWorker"
	"bot/internal/global"
	"bot/internal/logger"
)

type BotT struct {
	config v1alpha1.BOTConfigT
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

	level, err := logger.GetLevel(botServer.config.APIService.LogLevel)
	if err != nil {
		log.Fatalf("unable to get api service loglevel: %s", err.Error())
	}

	botServer.log = logger.NewLogger(context.Background(), level, map[string]any{
		"service":   "bot",
		"component": "none",
	})

	botServer.APIService = apiService.NewApiService(&botServer.config)

	botServer.ObjectWorker, err = objectWorker.NewObjectWorker(&botServer.config)
	if err != nil {
		return botServer, err
	}

	botServer.DatabaseWorker, err = databaseWorker.NewDatabaseWorker(&botServer.config)
	if err != nil {
		return botServer, err
	}

	botServer.HashringWorker = hashringWorker.NewHashringWorker(&botServer.config)

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
	log.Printf("executing shutdown actions with signal '%s'", sig.String())

	b.APIService.Shutdown()
	b.DatabaseWorker.Shutdown()
	b.ObjectWorker.Shutdown()
	b.HashringWorker.Shutdown()

	done <- true
}
