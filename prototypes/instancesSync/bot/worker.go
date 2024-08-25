package bot

import (
	"log"
	"time"
)

type WorkerT struct {
	// Instances map[string]ServerT
}

// WORKER Functions

func (b *BotT) workerFlow() {
	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestMap := TransferRequestPool.GetTransferRequestMap()

		for itemPath, request := range transferRequestMap {
			_ = request
			serverName := HashRing.GetNode(itemPath)

			if serverName != b.Server.Name {
				// send transfer request to owner
				log.Printf("moving '%s' transfer request from '%s' to '%s'", itemPath, b.Server.Name, serverName)

				// remove request from pool
				TransferRequestPool.RemoveTransferRequest(itemPath)
				continue
			}

			// process transfer request
			log.Printf("process '%s' transfer request in '%s'", itemPath, b.Server.Name)

			// remove request from pool
			TransferRequestPool.RemoveTransferRequest(itemPath)
		}

		time.Sleep(5 * time.Second) // REMOVE THIS IN THE END
	}
}

func (b *BotT) InitWorker() {
	go b.workerFlow()
}
