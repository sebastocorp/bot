package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"prototypes/instancesSync/bot"
)

func main() {
	botServer, err := bot.NewBotServer()
	if err != nil {
		log.Fatalf("unable to init bot server: %s", err.Error())
	}

	// create channels to manage shutdown actions
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	shutdownActionsDone := make(chan bool, 1)
	go botServer.ShutdownActions(shutdownActionsDone, signals)

	// check host is added in load balancer
	botServer.CheckOwnHost()

	// Init bot server
	botServer.InitSynchronizer()
	botServer.InitWorker()
	botServer.InitAPI()

	<-shutdownActionsDone
	log.Printf("shutdown of '%s' server", botServer.Server.Name)
}
