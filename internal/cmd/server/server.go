package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"bot/internal/bot"
	"bot/internal/global"
	"bot/internal/logger"
)

type serverT struct {
	context context.Context
	flags   serverFlagsT
}

func (d *serverT) flow() {
	botServer, err := bot.NewBotServer(d.flags.config)
	if err != nil {
		logger.Logger.Fatalf("unable to init bot server: %s", err.Error())
	}

	// create channels to manage shutdown actions
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	shutdownActionsDone := make(chan bool, 1)
	go botServer.ShutdownActions(shutdownActionsDone, signals)

	// check host is added in load balancer
	botServer.CheckOwnHost()
	logger.Logger.Infof("found '%s' own host in '%s' proxy host resolution", global.Config.APIService.Address, global.Config.HashRingWorker.Proxy)

	// Init bot server
	botServer.HashRingWorker.InitWorker()
	botServer.ObjectWorker.InitWorker()
	botServer.DatabaseWorker.InitWorker()
	botServer.APIService.InitAPI()

	<-shutdownActionsDone
}
