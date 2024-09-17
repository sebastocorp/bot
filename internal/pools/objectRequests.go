package pools

import (
	"maps"
	"sync"

	"bot/api/v1alpha1"
)

type ObjectRequestPoolT struct {
	mu       sync.Mutex
	requests map[string]v1alpha1.TransferRequestT
}

func NewTransferRequestPool() ObjectRequestPoolT {
	return ObjectRequestPoolT{
		requests: map[string]v1alpha1.TransferRequestT{},
	}
}

// REQUEST POOL FUNCTIONS

func (pool *ObjectRequestPoolT) GetPool() (result map[string]v1alpha1.TransferRequestT) {
	result = map[string]v1alpha1.TransferRequestT{}

	pool.mu.Lock()
	maps.Copy(result, pool.requests)
	pool.mu.Unlock()

	return result
}

func (pool *ObjectRequestPoolT) AddRequest(transfer v1alpha1.TransferRequestT) {
	pool.mu.Lock()
	pool.requests[transfer.To.ObjectPath] = transfer
	pool.mu.Unlock()
}

func (pool *ObjectRequestPoolT) RemoveRequest(key string) {
	pool.mu.Lock()
	delete(pool.requests, key)
	pool.mu.Unlock()
}
