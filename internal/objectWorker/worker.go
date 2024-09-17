package objectWorker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/objectStorage"
	"bot/internal/pools"
)

type ObjectWorkerT struct {
	ObjectManager objectStorage.ManagerT
}

// WORKER Functions

func (o *ObjectWorkerT) executeTransferRequest(request pools.TransferT) (err error) {
	// check if destination object already exist
	destInfo, err := o.ObjectManager.S3ObjectExist(request.To)
	if err != nil {
		return err
	}
	request.To.Info = destInfo

	if !destInfo.Exist {
		// check if source object already exist
		sourceInfo, err := o.ObjectManager.GCSObjectExist(request.From)
		if err != nil {
			return err
		}

		if !sourceInfo.Exist {
			err = fmt.Errorf("object '%s' NOT found in '%s' source bucket", request.From.ObjectPath, request.From.BucketName)
			return err
		}

		_, err = o.ObjectManager.TransferObjectFromGCSToS3(request.From, request.To)
		if err != nil {
			return err
		}

		request.To.Info = sourceInfo
	}

	global.DatabaseRequestPool.AddRequest(pools.DatabaseRequestT{
		BucketName: request.To.BucketName,
		ObjectPath: request.To.ObjectPath,
		MD5:        request.To.Info.MD5,
	})

	// // Get the object from the database
	// _, occurrences, err := o.DatabaseManager.GetObject(request.To)
	// if err != nil {
	// 	return err
	// }

	// // Insert the object in the database
	// if occurrences == 0 {
	// 	err = o.DatabaseManager.InsertObject(request.To)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return err
}

func (o *ObjectWorkerT) moveTransferRequest(serverName string, request pools.TransferT) (err error) {
	pool := global.ServerInstancesPool.GetPool()
	serverToSend := pools.ServerT{}
	for _, server := range pool {
		if server.Name == serverName {
			serverToSend = server
			break
		}
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	http.DefaultClient.Timeout = 100 * time.Millisecond
	// requestURL := fmt.Sprintf("http://%s:%s/transfer", serverToSend.Address, o.Config.Port)
	requestURL := fmt.Sprintf("http://%s:%s/transfer", serverToSend.Address, "8080")
	respBody, err := http.Post(requestURL, global.HeaderContentTypeAppJson, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	defer respBody.Body.Close()

	return err
}

func (o *ObjectWorkerT) processTransferRequest(wg *sync.WaitGroup, itemPath string, request pools.TransferT) {
	defer wg.Done()

	logger.Logger.Infof("process '%s' transfer request '%v'", itemPath, request)

	if global.ServerConfig.HashRingWorker.Enabled {
		serverName := global.HashRing.GetNode(itemPath)

		if serverName != global.ServerConfig.Name {
			// send transfer request to owner
			logger.Logger.Infof("moving '%s' transfer request to '%s'", itemPath, serverName)

			err := o.moveTransferRequest(serverName, request)
			if err == nil {
				return
			}

			logger.Logger.Errorf("unable to move '%s' transfer request to '%s': %s", itemPath, serverName, err.Error())
		}
	}

	logger.Logger.Infof("execute '%s' transfer request", itemPath)
	err := o.executeTransferRequest(request)
	if err != nil {
		logger.Logger.Errorf("unable to execute '%s' transfer request: %s", itemPath, err.Error())
	} else {
		logger.Logger.Infof("success executing '%s' transfer request", itemPath)
	}
}

func (o *ObjectWorkerT) workerFlow() {
	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestMap := global.TransferRequestPool.GetTransferRequestMap()

		wg := sync.WaitGroup{}
		count := 0
		for itemPath, request := range transferRequestMap {
			wg.Add(1)

			go o.processTransferRequest(&wg, itemPath, request)

			// remove request from pool
			global.TransferRequestPool.RemoveTransferRequest(itemPath)

			if count++; count >= global.ServerConfig.ObjectWorker.ParallelRequests {
				break
			}
		}

		wg.Wait()
	}
}

func (o *ObjectWorkerT) InitWorker() {
	go o.workerFlow()
}
