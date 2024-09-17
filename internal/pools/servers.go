package pools

import (
	"maps"
	"sync"
)

type ServerInstancesPoolT struct {
	mu      sync.Mutex
	servers map[string]ServerT
}

type ServerT struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	URL     string `json:"url"`
}

func NewServerInstancesPool() ServerInstancesPoolT {
	return ServerInstancesPoolT{
		servers: map[string]ServerT{},
	}
}

// SERVER POOL FUNCTIONS

func (pool *ServerInstancesPoolT) GetPool() (result map[string]ServerT) {
	result = map[string]ServerT{}

	pool.mu.Lock()
	maps.Copy(result, pool.servers)
	pool.mu.Unlock()

	return result
}

func (pool *ServerInstancesPoolT) GetServersList() (result []ServerT) {
	servers := pool.GetPool()
	result = []ServerT{}

	for _, server := range servers {
		result = append(result, server)
	}

	return result
}

func (pool *ServerInstancesPoolT) AddServers(servers []ServerT) {
	pool.mu.Lock()
	for _, server := range servers {
		pool.servers[server.Address] = server
	}
	pool.mu.Unlock()
}

func (pool *ServerInstancesPoolT) RemoveServers(servers []ServerT) {
	pool.mu.Lock()
	for _, server := range servers {
		delete(pool.servers, server.Address)
	}
	pool.mu.Unlock()
}
