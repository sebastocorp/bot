package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"bot/internal/logger"
)

// WORKER Functions

func (b *BotT) executeTransferRequest(request TransferT) (err error) {
	// check if destination object already exist
	destInfo, err := b.ObjectManager.S3ObjectExist(request.To)
	if err != nil {
		return err
	}
	request.To.Info = destInfo

	if !destInfo.Exist {
		// check if source object already exist
		sourceInfo, err := b.ObjectManager.GCSObjectExist(request.From)
		if err != nil {
			return err
		}

		if !sourceInfo.Exist {
			err = fmt.Errorf("object '%s' NOT found in '%s' source bucket", request.From.ObjectPath, request.From.BucketName)
			return err
		}

		_, err = b.ObjectManager.TransferObjectFromGCSToS3(request.From, request.To)
		if err != nil {
			return err
		}

		request.To.Info = sourceInfo
	}

	// Get the object from the database
	_, occurrences, err := b.DatabaseManager.GetObject(request.To)
	if err != nil {
		return err
	}

	// Insert the object in the database
	if occurrences == 0 {
		err = b.DatabaseManager.InsertObject(request.To)
		if err != nil {
			return err
		}
	}

	return err
}

func (b *BotT) moveTransferRequest(serverName string, request TransferT) (err error) {
	pool := ServerInstancesPool.GetPool()
	serverToSend := ServerT{}
	for _, server := range pool {
		if server.Name == serverName {
			serverToSend = server
			break
		}
	}

	body := apiTransferRequestT{
		Transfer: request,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	http.DefaultClient.Timeout = 200 * time.Millisecond
	requestURL := fmt.Sprintf("http://%s:%s/transfer", serverToSend.Address, b.API.Port)
	respBody, err := http.Post(requestURL, headerContentTypeAppJson, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	defer respBody.Body.Close()

	return err
}

func (b *BotT) processTransferRequest(wg *sync.WaitGroup, itemPath string, request TransferT) {
	defer wg.Done()

	logger.Logger.Infof("process '%s' transfer request '%v' in '%s'", itemPath, request, b.Server.Name)
	serverName := HashRing.GetNode(itemPath)

	if serverName != b.Server.Name {
		// send transfer request to owner
		logger.Logger.Infof("moving '%s' transfer request from '%s' to '%s'", itemPath, b.Server.Name, serverName)

		err := b.moveTransferRequest(serverName, request)
		if err == nil {
			return
		}

		logger.Logger.Errorf("unable to move '%s' transfer request to '%s': %s", itemPath, serverName, err.Error())
	}

	logger.Logger.Infof("execute '%s' transfer request in '%s'", itemPath, b.Server.Name)
	err := b.executeTransferRequest(request)
	if err != nil {
		logger.Logger.Errorf("unable to execute '%s' transfer request in '%s': %s", itemPath, b.Server.Name, err.Error())
	} else {
		logger.Logger.Infof("success executing '%s' transfer request in '%s'", itemPath, b.Server.Name)
	}
}

func (b *BotT) workerFlow() {
	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestMap := TransferRequestPool.GetTransferRequestMap()

		wg := sync.WaitGroup{}
		count := 0
		for itemPath, request := range transferRequestMap {
			wg.Add(1)

			go b.processTransferRequest(&wg, itemPath, request)

			// remove request from pool
			TransferRequestPool.RemoveTransferRequest(itemPath)

			if count++; count >= b.ParallelRequests {
				break
			}
		}

		wg.Wait()
	}
}

func (b *BotT) InitWorker() {
	go b.workerFlow()
}
