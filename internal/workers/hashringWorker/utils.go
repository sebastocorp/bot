package hashringWorker

import (
	"net"
	"slices"
	"time"

	"bot/internal/global"
	"bot/internal/logger"
)

func (h *HashRingWorkerT) CheckOwnHost() {

	found := false
	for !found {
		time.Sleep(4 * time.Second)

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