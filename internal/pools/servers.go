package pools

import (
	"maps"
	"sync"

	"bot/api/v1alpha1"
)

type ServerInstancesPoolT struct {
	mu      sync.Mutex
	servers map[string]v1alpha1.ServerT
}

func NewServerPool() *ServerInstancesPoolT {
	return &ServerInstancesPoolT{
		servers: map[string]v1alpha1.ServerT{},
	}
}

func (pool *ServerInstancesPoolT) GetPool() (result map[string]v1alpha1.ServerT) {
	result = map[string]v1alpha1.ServerT{}

	pool.mu.Lock()
	maps.Copy(result, pool.servers)
	pool.mu.Unlock()

	return result
}

func (pool *ServerInstancesPoolT) GetServersList() (result []v1alpha1.ServerT) {
	servers := pool.GetPool()
	result = []v1alpha1.ServerT{}

	for _, server := range servers {
		result = append(result, server)
	}

	return result
}

func (pool *ServerInstancesPoolT) AddServers(servers []v1alpha1.ServerT) {
	pool.mu.Lock()
	for _, server := range servers {
		pool.servers[server.Address] = server
	}
	pool.mu.Unlock()
}

func (pool *ServerInstancesPoolT) RemoveServers(servers []v1alpha1.ServerT) {
	pool.mu.Lock()
	for _, server := range servers {
		delete(pool.servers, server.Address)
	}
	pool.mu.Unlock()
}
