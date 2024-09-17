package pools

import (
	"maps"
	"sync"

	"bot/api/v1alpha1"
)

type DatabaseRequestPoolT struct {
	mu       sync.Mutex
	requests map[string]v1alpha1.DatabaseRequestT
}

func NewDatabaseRequestPool() *DatabaseRequestPoolT {
	return &DatabaseRequestPoolT{
		requests: map[string]v1alpha1.DatabaseRequestT{},
	}
}

// SERVER POOL FUNCTIONS

func (pool *DatabaseRequestPoolT) GetPool() (result map[string]v1alpha1.DatabaseRequestT) {
	result = map[string]v1alpha1.DatabaseRequestT{}

	pool.mu.Lock()
	maps.Copy(result, pool.requests)
	pool.mu.Unlock()

	return result
}

func (pool *DatabaseRequestPoolT) AddRequest(request v1alpha1.DatabaseRequestT) {
	pool.mu.Lock()
	pool.requests[request.ObjectPath] = request
	pool.mu.Unlock()
}

func (pool *DatabaseRequestPoolT) RemoveRequest(key string) {
	pool.mu.Lock()
	delete(pool.requests, key)
	pool.mu.Unlock()
}

func (pool *DatabaseRequestPoolT) RemoveRequests(requests []v1alpha1.DatabaseRequestT) {
	pool.mu.Lock()
	for _, req := range pool.requests {
		delete(pool.requests, req.ObjectPath)
	}
	pool.mu.Unlock()
}
