package hashringWorker

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/hashring"
)

type HashRingWorkerT struct {
}

func (h *HashRingWorkerT) discoverServerAddresses() (instancesAddrs []string, err error) {
	discoveredHosts, err := net.LookupHost(global.Config.HashRingWorker.Proxy)
	if err != nil {
		return instancesAddrs, err
	}

	for _, dHost := range discoveredHosts {
		if dHost != global.Config.APIService.Address {
			instancesAddrs = append(instancesAddrs, dHost)
		}
	}

	return instancesAddrs, err
}

func (h *HashRingWorkerT) checkHealth(address string) (err error) {
	requestURL := fmt.Sprintf("http://%s:%s%s",
		address,
		global.Config.APIService.Port,
		global.EndpointHealth,
	)
	res, err := http.Get(requestURL)
	if err != nil {
		return err
	}

	if res != nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("ready endpoint return not OK status")
	}

	return err
}

func (h *HashRingWorkerT) getServersInfo(addrsAdded []string) (result []v1alpha1.ServerT) {
	for _, address := range addrsAdded {
		err := h.checkHealth(address)
		if err != nil {
			logger.Logger.Errorf("error checking api of instance with address '%s': %s", address, err.Error())
			continue
		}

		requestURL := fmt.Sprintf("http://%s:%s%s",
			address,
			global.Config.APIService.Port,
			global.EndpointInfo,
		)
		res, err := http.Get(requestURL)
		if err != nil {
			logger.Logger.Errorf("error getting info of instance with address '%s': %s", address, err.Error())
			continue
		}

		resBodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Logger.Errorf("error reading info request body of instance with address '%s': %s", address, err.Error())
			res.Body.Close()
			continue
		}
		res.Body.Close()

		server := v1alpha1.ServerT{}
		err = json.Unmarshal(resBodyBytes, &server)
		if err != nil {
			logger.Logger.Errorf("error parsing info request body of instance with address '%s': %s", address, err.Error())
			continue
		}

		result = append(result, server)
	}

	return result
}

func (h *HashRingWorkerT) getServersPoolChanges(currentServersAddrsList []string) (added []v1alpha1.ServerT, removed []v1alpha1.ServerT) {
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

func (h *HashRingWorkerT) flow() {
	for {
		time.Sleep(2 * time.Second)

		currentServersAddrsList, err := h.discoverServerAddresses()
		if err != nil {
			logger.Logger.Errorf("unable to discover current servers in '%s' proxy host: %s", global.Config.HashRingWorker.Proxy, err.Error())
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

func (h *HashRingWorkerT) InitWorker() {
	if global.Config.HashRingWorker.Enabled {
		// check host is added in load balancer
		h.CheckOwnHost()
		logger.Logger.Infof("found '%s' own host in '%s' proxy host resolution", global.Config.APIService.Address, global.Config.HashRingWorker.Proxy)

		global.HashRing = hashring.NewHashRing(global.Config.HashRingWorker.VNodes)

		global.HashRing.AddNodes([]string{global.Config.Name})

		global.ServerState.SetHashringReady()

		go h.flow()
	} else {
		global.ServerState.SetHashringReady()
	}
}
