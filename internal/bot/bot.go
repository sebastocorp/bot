package bot

import (
	"log"
	"os"
	"time"

	"bot/internal/components/apiService"
	"bot/internal/components/databaseWorker"
	"bot/internal/components/hashringWorker"
	"bot/internal/components/objectWorker"
	"bot/internal/global"
)

type BotT struct {
	APIService     *apiService.APIServiceT
	ObjectWorker   *objectWorker.ObjectWorkerT
	DatabaseWorker *databaseWorker.DatabaseWorkerT
	HashRingWorker *hashringWorker.HashRingWorkerT
}

// BOT SERVER FUNCTIONS

func NewBotServer(configFilepath string) (botServer *BotT, err error) {
	botServer = &BotT{}

	botConfig, err := parseConfig(configFilepath)
	if err != nil {
		return botServer, err
	}

	//--------------------------------------------------------------
	// CHECK API CONFIG
	//--------------------------------------------------------------
	botServer.APIService = apiService.NewApiService(botConfig.APIService)

	//--------------------------------------------------------------
	// CHECK OBJECT CONFIG
	//--------------------------------------------------------------
	botServer.ObjectWorker, err = objectWorker.NewObjectWorker(botConfig.ObjectWorker)

	//--------------------------------------------------------------
	// CHECK DATABASE CONFIG
	//--------------------------------------------------------------
	botServer.DatabaseWorker, err = databaseWorker.NewDatabaseWorker(botConfig.DatabaseWorker)

	//--------------------------------------------------------------
	// CHECK HASHRING CONFIG
	//--------------------------------------------------------------
	botServer.HashRingWorker = hashringWorker.NewHashringWorker(botConfig.HashRingWorker)

	return botServer, err
}

func (b *BotT) Run() {
	// Init bot server
	b.HashRingWorker.InitWorker()
	b.ObjectWorker.InitWorker()
	b.DatabaseWorker.Run()
	b.APIService.Run()

	for !global.ServerState.IsReady() {
		log.Printf("waiting for bot server ready...")
		time.Sleep(5 * time.Second)
	}

	log.Printf("bot server is ready")
}

func (b *BotT) ShutdownActions(done chan bool, signal chan os.Signal) {
	sig := <-signal
	log.Printf("executing shutdown actions with signal '%s'", sig.String())

	b.APIService.Shutdown()
	b.DatabaseWorker.Shutdown()

	done <- true
}
