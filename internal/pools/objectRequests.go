package pools

import (
	"fmt"
	"maps"
	"sync"

	"bot/api/v1alpha1"
)

type ObjectRequestPoolT struct {
	mu       sync.Mutex
	requests map[string]ObjectRequestT
}

type ObjectRequestT struct {
	Object v1alpha1.ObjectT
}

func NewTransferRequestPool() ObjectRequestPoolT {
	return ObjectRequestPoolT{
		requests: map[string]ObjectRequestT{},
	}
}

// REQUEST POOL FUNCTIONS

func (pool *ObjectRequestPoolT) GetPool() (result map[string]ObjectRequestT) {
	result = map[string]ObjectRequestT{}

	pool.mu.Lock()
	maps.Copy(result, pool.requests)
	pool.mu.Unlock()

	return result
}

func (pool *ObjectRequestPoolT) AddRequest(transfer ObjectRequestT) {
	pool.mu.Lock()
	pool.requests[transfer.Object.Path] = transfer
	pool.mu.Unlock()
}

func (pool *ObjectRequestPoolT) RemoveRequest(key string) {
	pool.mu.Lock()
	delete(pool.requests, key)
	pool.mu.Unlock()
}

func (pool *ObjectRequestPoolT) RemoveRequests(requests []ObjectRequestT) {
	pool.mu.Lock()
	for _, req := range pool.requests {
		delete(pool.requests, req.Object.Path)
	}
	pool.mu.Unlock()
}

func (or *ObjectRequestT) String() string {
	return fmt.Sprintf("{bucket: '%s', object: '%s'}", or.Object.Bucket, or.Object.Path)
}
