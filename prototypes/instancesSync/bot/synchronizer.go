package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func (b *BotT) getServersInfo(addrsAdded []string) (result []ServerT, err error) {
	for _, address := range addrsAdded {
		requestURL := fmt.Sprintf("http://%s:%s/instance", address, b.API.Port)

		err = b.checkAPI(requestURL + "/ready")
		if err != nil {
			log.Printf("api '%s' not ready: %s", requestURL, err.Error())
			continue
		}

		res, err := http.Get(requestURL + "/info")
		if err != nil {
			continue
		}

		resBodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			continue
		}
		res.Body.Close()

		body := resInfoBodyT{}
		err = json.Unmarshal(resBodyBytes, &body)
		if err != nil {
			continue
		}

		result = append(result, ServerT{
			Name:    body.Server.Name,
			Address: address,
		})
	}

	return result, err
}

func (b *BotT) getServersPoolChanges(currentServersAddrsList []string) (addrsAdded []string, removed []ServerT) {
	storedPool := ServerInstancesPool.GetPool()

	for _, server := range storedPool {
		if !slices.Contains(currentServersAddrsList, server.Address) {
			removed = append(removed, server)
		}
	}

	for _, addr := range currentServersAddrsList {
		if _, ok := storedPool[addr]; !ok {
			addrsAdded = append(addrsAdded, addr)
		}
	}

	return addrsAdded, removed
}

func (b *BotT) synchronizerFlow() {
	for {
		time.Sleep(2 * time.Second)

		currentServersAddrsList, err := b.discoverServers()
		if err != nil {
			log.Printf("unable to discover current servers in '%s' proxy host: %s", b.ProxyHost, err.Error())
		}

		addrsAdded, serversRemoved := b.getServersPoolChanges(currentServersAddrsList)

		if len(addrsAdded) != 0 || len(serversRemoved) != 0 {
			log.Printf("update servers pool and compute hashring in '%s'", b.Server.Name)

			serversAdded, err := b.getServersInfo(addrsAdded)
			if err != nil {
				log.Printf("unable to get new servers instances: %s", err.Error())
				continue
			}

			log.Printf("adding '%s' servers in '%s' pool", serversAdded, b.Server.Name)
			ServerInstancesPool.AddServers(serversAdded)
			log.Printf("removing '%s' servers in '%s' pool", serversRemoved, b.Server.Name)
			ServerInstancesPool.RemoveServers(serversRemoved)

			// GENERATE HASH RING

			added := []string{}
			removed := []string{}

			if !HashRing.InHashRing(b.Server.Name) {
				added = append(added, b.Server.Name)
			}

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
	go b.synchronizerFlow()
}
