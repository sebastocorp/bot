package bot

import (
	"bot/internal/logger"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"time"
)

func (b *BotT) discoverServers() (instancesAddrs []string, err error) {
	discoveredHosts, err := net.LookupHost(b.ProxyHost)
	if err != nil {
		return instancesAddrs, err
	}

	for _, dHost := range discoveredHosts {
		if dHost != b.Server.Address {
			instancesAddrs = append(instancesAddrs, dHost)
		}
	}

	return instancesAddrs, err
}

func (b *BotT) checkAPI(address string) (err error) {
	requestURL := fmt.Sprintf("http://%s:%s/status", address, b.API.Port)
	res, err := http.Get(requestURL)
	if err != nil {
		return err
	}

	if res != nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("ready endpoint return not OK status")
	}

	return err
}

func (b *BotT) getServersInfo(addrsAdded []string) (result []ServerT) {
	for _, address := range addrsAdded {
		err := b.checkAPI(address)
		if err != nil {
			logger.Logger.Errorf("error checking api of instance with address '%s': %s", address, err.Error())
			continue
		}

		requestURL := fmt.Sprintf("http://%s:%s/info", address, b.API.Port)
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

		body := apiInfoRequestT{}
		err = json.Unmarshal(resBodyBytes, &body)
		if err != nil {
			logger.Logger.Errorf("error parsing info request body of instance with address '%s': %s", address, err.Error())
			continue
		}

		result = append(result, ServerT{
			Name:    body.Server.Name,
			Address: address,
		})
	}

	return result
}

func (b *BotT) getServersPoolChanges(currentServersAddrsList []string) (added []ServerT, removed []ServerT) {
	storedPool := ServerInstancesPool.GetPool()

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

	added = b.getServersInfo(addrsAdded)

	return added, removed
}

func (b *BotT) synchronizerFlow() {
	if !HashRing.InHashRing(b.Server.Name) {
		HashRing.AddNodes([]string{b.Server.Name})
	}

	for {
		time.Sleep(2 * time.Second)

		currentServersAddrsList, err := b.discoverServers()
		if err != nil {
			logger.Logger.Errorf("unable to discover current servers in '%s' proxy host: %s", b.ProxyHost, err.Error())
		}

		serversAdded, serversRemoved := b.getServersPoolChanges(currentServersAddrsList)

		if len(serversAdded) != 0 || len(serversRemoved) != 0 {
			logger.Logger.Infof("update servers pool and compute hashring in '%s'", b.Server.Name)

			ServerInstancesPool.AddServers(serversAdded)
			ServerInstancesPool.RemoveServers(serversRemoved)

			serversList := ServerInstancesPool.GetServersList()
			logger.Logger.Infof("current servers in '%s' instance: %v", b.Server.Name, serversList)

			// GENERATE HASH RING

			added := []string{}
			removed := []string{}

			for _, server := range serversAdded {
				added = append(added, server.Name)
			}

			for _, server := range serversRemoved {
				removed = append(removed, server.Name)
			}

			HashRing.RemoveNodes(removed)
			HashRing.AddNodes(added)
		}
	}
}

func (b *BotT) InitSynchronizer() {
	if b.UseHashRing {
		go b.synchronizerFlow()
	}
}
