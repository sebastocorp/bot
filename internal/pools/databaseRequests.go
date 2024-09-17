package pools

import (
	"maps"
	"sync"
)

type DatabaseRequestPoolT struct {
	mu       sync.Mutex
	requests map[string]DatabaseRequestT
}

type DatabaseRequestT struct {
	BucketName string `json:"bucket"`
	ObjectPath string `json:"path"`
	MD5        string `json:"md5"`
}

func NewDatabaseRequestPool() *DatabaseRequestPoolT {
	return &DatabaseRequestPoolT{
		requests: map[string]DatabaseRequestT{},
	}
}

// SERVER POOL FUNCTIONS

func (pool *DatabaseRequestPoolT) GetPool() (result map[string]DatabaseRequestT) {
	result = map[string]DatabaseRequestT{}

	pool.mu.Lock()
	maps.Copy(result, pool.requests)
	pool.mu.Unlock()

	return result
}

func (pool *DatabaseRequestPoolT) AddRequest(request DatabaseRequestT) {
	pool.mu.Lock()
	pool.requests[request.ObjectPath] = request
	pool.mu.Unlock()
}

func (pool *DatabaseRequestPoolT) RemoveRequest(key string) {
	pool.mu.Lock()
	delete(pool.requests, key)
	pool.mu.Unlock()
}

func (pool *DatabaseRequestPoolT) RemoveRequests(requests []DatabaseRequestT) {
	pool.mu.Lock()
	for _, req := range pool.requests {
		delete(pool.requests, req.ObjectPath)
	}
	pool.mu.Unlock()
}
