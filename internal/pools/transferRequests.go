package pools

import (
	"maps"
	"sync"

	"bot/internal/objectStorage"
)

type TransferRequestPoolT struct {
	mu       sync.Mutex
	requests map[string]TransferT
}

type TransferT struct {
	From objectStorage.ObjectT `json:"from"`
	To   objectStorage.ObjectT `json:"to"`
}

func NewTransferRequestPool() TransferRequestPoolT {
	return TransferRequestPoolT{
		requests: map[string]TransferT{},
	}
}

// REQUEST POOL FUNCTIONS

func (pool *TransferRequestPoolT) AddTransferRequest(transfer TransferT) {
	pool.mu.Lock()
	pool.requests[transfer.To.ObjectPath] = transfer
	pool.mu.Unlock()
}

func (pool *TransferRequestPoolT) RemoveTransferRequest(transferKey string) {
	pool.mu.Lock()
	delete(pool.requests, transferKey)
	pool.mu.Unlock()
}

func (pool *TransferRequestPoolT) GetTransferRequestMap() (result map[string]TransferT) {
	result = map[string]TransferT{}

	pool.mu.Lock()
	maps.Copy(result, pool.requests)
	pool.mu.Unlock()

	return result
}
