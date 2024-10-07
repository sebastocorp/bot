package hashringWorker

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"time"

	"bot/api/v1alpha2"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/hashring"
	"bot/internal/pools"
)

type HashringWorkerT struct {
	config *v1alpha2.BOTConfigT
	log    logger.LoggerT

	hashring           *hashring.HashRingT
	serverInstancePool *pools.ServerInstancesPoolT
}

func NewHashringWorker(config *v1alpha2.BOTConfigT, serverPool *pools.ServerInstancesPoolT) (hw *HashringWorkerT) {
	hw = &HashringWorkerT{
		config:             config,
		serverInstancePool: serverPool,
	}

	return hw
}

func (hw *HashringWorkerT) Run() {
	if hw.config.HashRingWorker.Enabled {
		// check host is added in load balancer
		hw.CheckOwnHost()
		// TODO: add config to get api address
		// hw.log.Info("found '%s' own host in '%s' proxy host resolution", "global.Config.APIService.Address", hw.config.HashRingWorker.Proxy)

		hw.hashring = hashring.NewHashRing(hw.config.HashRingWorker.VNodes)

		hw.hashring.AddNodes([]string{"global.Config.Name"})

		global.ServerState.SetHashringReady()

		go hw.flow()
	}

	global.ServerState.SetHashringReady()
}

func (hw *HashringWorkerT) Shutdown() {

}

func (hw *HashringWorkerT) discoverServerAddresses() (instancesAddrs []string, err error) {
	discoveredHosts, err := net.LookupHost(hw.config.HashRingWorker.Proxy)
	if err != nil {
		return instancesAddrs, err
	}

	for _, dHost := range discoveredHosts {
		// TODO: add config to get api
		if dHost != "global.Config.APIService.Address" {
			instancesAddrs = append(instancesAddrs, dHost)
		}
	}

	return instancesAddrs, err
}

func (h *HashringWorkerT) checkHealth(address string) (err error) {
	// TODO: add config to get api
	requestURL := fmt.Sprintf("http://%s:%s%s",
		address,
		"global.Config.APIService.Port",
		global.EndpointHealthz,
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

func (hw *HashringWorkerT) getServersInfo(addrsAdded []string) (result []pools.ServerT) {
	for _, address := range addrsAdded {
		err := hw.checkHealth(address)
		if err != nil {
			// hw.log.Error("error checking api of instance with address '%s': %s", address, err.Error())
			continue
		}

		requestURL := fmt.Sprintf("http://%s:%s%s",
			address,
			"global.Config.APIService.Port",
			global.EndpointInfo,
		)
		res, err := http.Get(requestURL)
		if err != nil {
			// logger.Logger.Errorf("error getting info of instance with address '%s': %s", address, err.Error())
			continue
		}

		resBodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			// logger.Logger.Errorf("error reading info request body of instance with address '%s': %s", address, err.Error())
			res.Body.Close()
			continue
		}
		res.Body.Close()

		server := pools.ServerT{}
		err = json.Unmarshal(resBodyBytes, &server)
		if err != nil {
			// logger.Logger.Errorf("error parsing info request body of instance with address '%s': %s", address, err.Error())
			continue
		}

		result = append(result, server)
	}

	return result
}

func (hw *HashringWorkerT) getServersPoolChanges(currentServersAddrsList []string) (added []pools.ServerT, removed []pools.ServerT) {
	storedPool := hw.serverInstancePool.GetPool()

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

	added = hw.getServersInfo(addrsAdded)

	return added, removed
}

func (hw *HashringWorkerT) flow() {
	for {
		time.Sleep(2 * time.Second)

		currentServersAddrsList, err := hw.discoverServerAddresses()
		if err != nil {
			// logger.Logger.Errorf("unable to discover current servers in '%s' proxy host: %s", global.Config.HashRingWorker.Proxy, err.Error())
		}

		serversAdded, serversRemoved := hw.getServersPoolChanges(currentServersAddrsList)

		if len(serversAdded) != 0 || len(serversRemoved) != 0 {
			// logger.Logger.Infof("update servers pool and compute hashring")

			hw.serverInstancePool.AddServers(serversAdded)
			hw.serverInstancePool.RemoveServers(serversRemoved)

			serversList := hw.serverInstancePool.GetServersList()
			_ = serversList
			// logger.Logger.Infof("current servers in hashring: %v", serversList)

			// GENERATE HASH RING

			added := []string{}
			removed := []string{}

			for _, server := range serversAdded {
				added = append(added, server.Name)
			}

			for _, server := range serversRemoved {
				removed = append(removed, server.Name)
			}

			hw.hashring.RemoveNodes(removed)
			hw.hashring.AddNodes(added)
		}
	}
}
