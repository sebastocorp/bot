package objectWorker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/objectStorage"
)

type ObjectWorkerT struct {
	ObjectManager objectStorage.ManagerT
}

// WORKER Functions

func (o *ObjectWorkerT) executeTransferRequest(request v1alpha1.TransferRequestT) (err error) {
	// check if destination object already exist
	destInfo, err := o.ObjectManager.S3ObjectExist(request.To)
	if err != nil {
		return err
	}
	request.To.Info = destInfo

	if !destInfo.Exist {
		sourceInfo, err := o.ObjectManager.TransferObjectFromGCSToS3(request.From, request.To)
		if err != nil {
			return err
		}

		request.To.Info = sourceInfo
	}

	global.DatabaseRequestPool.AddRequest(v1alpha1.DatabaseRequestT{
		BucketName: request.To.BucketName,
		ObjectPath: request.To.ObjectPath,
		MD5:        request.To.Info.MD5,
	})

	return err
}

func (o *ObjectWorkerT) moveTransferRequest(serverName string, request v1alpha1.TransferRequestT) (err error) {
	pool := global.ServerInstancesPool.GetPool()
	serverToSend := v1alpha1.ServerT{}
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
	requestURL := fmt.Sprintf("http://%s:%s/transfer", serverToSend.Address, global.Config.APIService.Port)
	respBody, err := http.Post(requestURL, global.HeaderContentTypeAppJson, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	defer respBody.Body.Close()

	return err
}

func (o *ObjectWorkerT) processTransferRequest(wg *sync.WaitGroup, itemPath string, request v1alpha1.TransferRequestT) {
	defer wg.Done()

	logger.Logger.Infof("process '%s' transfer request '%v'", itemPath, request)

	if global.Config.HashRingWorker.Enabled {
		serverName := global.HashRing.GetNode(itemPath)

		if serverName != global.Config.Name {
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
		transferRequestMap := global.TransferRequestPool.GetPool()

		wg := sync.WaitGroup{}
		count := 0
		for itemPath, request := range transferRequestMap {
			wg.Add(1)

			go o.processTransferRequest(&wg, itemPath, request)

			// remove request from pool
			global.TransferRequestPool.RemoveRequest(itemPath)

			if count++; count >= global.Config.ObjectWorker.ParallelRequests {
				break
			}
		}

		wg.Wait()
	}
}

func (o *ObjectWorkerT) InitWorker() {
	global.ServerState.SetObjectReady()
	go o.workerFlow()
}
