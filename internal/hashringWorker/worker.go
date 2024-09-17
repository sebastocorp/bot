package hashringWorker

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"time"

	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/pools"
)

type HashRingWorkerT struct {
}

func (h *HashRingWorkerT) discoverServers() (instancesAddrs []string, err error) {
	discoveredHosts, err := net.LookupHost(global.ServerConfig.HashRingWorker.Proxy)
	if err != nil {
		return instancesAddrs, err
	}

	for _, dHost := range discoveredHosts {
		if dHost != global.ServerConfig.APIService.Address {
			instancesAddrs = append(instancesAddrs, dHost)
		}
	}

	return instancesAddrs, err
}

func (h *HashRingWorkerT) checkAPI(address string) (err error) {
	requestURL := fmt.Sprintf("%s/health", global.ServerReference.URL)
	res, err := http.Get(requestURL)
	if err != nil {
		return err
	}

	if res != nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("ready endpoint return not OK status")
	}

	return err
}

func (h *HashRingWorkerT) getServersInfo(addrsAdded []string) (result []pools.ServerT) {
	for _, address := range addrsAdded {
		err := h.checkAPI(address)
		if err != nil {
			logger.Logger.Errorf("error checking api of instance with address '%s': %s", address, err.Error())
			continue
		}

		requestURL := fmt.Sprintf("%s/info", global.ServerReference.URL)
		res, err := http.Get(requestURL)
		if err != nil {
			logger.Logger.Errorf("error getting info of instance with address '%s': %s", address, err.Error())
			continue
		}

		resBodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Logger.Errorf("error reading info request body of instance with address '%s': %s", address, err.Error())
			continue
		}
		res.Body.Close()

		server := pools.ServerT{}
		err = json.Unmarshal(resBodyBytes, &server)
		if err != nil {
			logger.Logger.Errorf("error parsing info request body of instance with address '%s': %s", address, err.Error())
			continue
		}

		result = append(result, server)
	}

	return result
}

func (h *HashRingWorkerT) getServersPoolChanges(currentServersAddrsList []string) (added []pools.ServerT, removed []pools.ServerT) {
	storedPool := global.ServerInstancesPool.GetPool()

	for _, server := range storedPool {
		if !slices.Contains(currentServersAddrsList, server.Address) {
			removed = append(removed, server)
		}
	}

	addrsAdded := []string{}
	for _, addr := range currentServersAddrsList {
		if _, ok := storedPool[addr]; !ok {
			addrsAdded = append(addrsAdded, addr)
		}
	}

	added = h.getServersInfo(addrsAdded)

	return added, removed
}

func (h *HashRingWorkerT) synchronizerFlow() {
	if !global.HashRing.InHashRing(global.ServerReference.Name) {
		global.HashRing.AddNodes([]string{global.ServerReference.Name})
	}

	for {
		time.Sleep(2 * time.Second)

		currentServersAddrsList, err := h.discoverServers()
		if err != nil {
			logger.Logger.Errorf("unable to discover current servers in '%s' proxy host: %s", global.ServerConfig.HashRingWorker.Proxy, err.Error())
		}

		serversAdded, serversRemoved := h.getServersPoolChanges(currentServersAddrsList)

		if len(serversAdded) != 0 || len(serversRemoved) != 0 {
			logger.Logger.Infof("update servers pool and compute hashring")

			global.ServerInstancesPool.AddServers(serversAdded)
			global.ServerInstancesPool.RemoveServers(serversRemoved)

			serversList := global.ServerInstancesPool.GetServersList()
			logger.Logger.Infof("current servers in hashring: %v", serversList)

			// GENERATE HASH RING

			added := []string{}
			removed := []string{}

			for _, server := range serversAdded {
				added = append(added, server.Name)
			}

			for _, server := range serversRemoved {
				removed = append(removed, server.Name)
			}

			global.HashRing.RemoveNodes(removed)
			global.HashRing.AddNodes(added)
		}
	}
}

func (h *HashRingWorkerT) InitSynchronizer() {
	if global.ServerConfig.HashRingWorker.Enabled {
		go h.synchronizerFlow()
	}
}
