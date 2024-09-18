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

func (o *ObjectWorkerT) processTransferRequest(wg *sync.WaitGroup, request v1alpha1.TransferRequestT) {
	defer wg.Done()

	logger.Logger.Infof("process object transfer request '%s'", request.String())

	if global.Config.HashRingWorker.Enabled {
		serverName := global.HashRing.GetNode(request.To.ObjectPath)

		if serverName != global.Config.Name {
			// send transfer request to owner
			logger.Logger.Infof("moving object transfer request '%s' to '%s'", request.String(), serverName)

			err := o.moveTransferRequest(serverName, request)
			if err == nil {
				return
			}

			logger.Logger.Errorf("unable to move object transfer request '%s' to '%s': %s", request.String(), serverName, err.Error())
		}
	}

	err := o.executeTransferRequest(request)
	if err != nil {
		logger.Logger.Errorf("unable to process object transfer request '%s': %s", request.String(), err.Error())
	} else {
		logger.Logger.Infof("success in process object transfer request '%s'", request.String())
	}
}

func (o *ObjectWorkerT) processRequestList(wg *sync.WaitGroup, requests []v1alpha1.TransferRequestT) {
	defer wg.Done()

	for _, request := range requests {
		logger.Logger.Infof("process object transfer request '%s'", request.String())

		if global.Config.HashRingWorker.Enabled {
			serverName := global.HashRing.GetNode(request.To.ObjectPath)

			if serverName != global.Config.Name {
				// send transfer request to owner
				logger.Logger.Infof("moving object transfer request '%s' to '%s'", request.String(), serverName)

				err := o.moveTransferRequest(serverName, request)
				if err == nil {
					return
				}

				logger.Logger.Errorf("unable to move object transfer request '%s' to '%s': %s", request.String(), serverName, err.Error())
			}
		}

		err := o.executeTransferRequest(request)
		if err != nil {
			logger.Logger.Errorf("unable to process object transfer request '%s': %s", request.String(), err.Error())
		} else {
			logger.Logger.Infof("success in process object transfer request '%s'", request.String())
		}
	}
}

func (o *ObjectWorkerT) flow() {
	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestPool := global.TransferRequestPool.GetPool()
		poolLen := len(transferRequestPool)
		if poolLen == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		wg := sync.WaitGroup{}
		currentThreads := 0
		for itemPath, request := range transferRequestPool {
			wg.Add(1)

			go o.processTransferRequest(&wg, request)
			global.TransferRequestPool.RemoveRequest(itemPath)

			if currentThreads++; currentThreads >= global.Config.ObjectWorker.MaxChildTheads {
				break
			}
		}

		logger.Logger.Infof("current object worker status {threads: '%d', pool_length: '%d'}", currentThreads, poolLen)
		wg.Wait()
	}
}

func (o *ObjectWorkerT) multiRequestFlow() {
	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestPool := global.TransferRequestPool.GetPool()

		poolLen := len(transferRequestPool)
		if poolLen == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		threadList := [][]v1alpha1.TransferRequestT{}
		requestList := []v1alpha1.TransferRequestT{}
		currentThreads := 0
		requestIndex := 0
		for _, request := range transferRequestPool {
			requestList = append(requestList, request)

			if requestIndex++; requestIndex >= global.Config.ObjectWorker.RequestsByChildThread {
				threadList = append(threadList, requestList)
				requestIndex = 0
				requestList = []v1alpha1.TransferRequestT{}

				if currentThreads++; currentThreads >= global.Config.ObjectWorker.MaxChildTheads {
					break
				}
			}
		}

		if len(requestList) > 0 {
			threadList = append(threadList, requestList)
			currentThreads++
		}

		wg := sync.WaitGroup{}
		for _, requests := range threadList {
			wg.Add(1)

			go o.processRequestList(&wg, requests)
			global.TransferRequestPool.RemoveRequests(requests)
		}

		logger.Logger.Infof("current object worker status {threads: '%d', pool_length: '%d'}", currentThreads, poolLen)
		wg.Wait()
	}
}

func (o *ObjectWorkerT) InitWorker() {
	global.ServerState.SetObjectReady()
	if global.Config.ObjectWorker.RequestsByChildThread > 0 {
		o.multiRequestFlow()
	} else {
		go o.flow()
	}
}
