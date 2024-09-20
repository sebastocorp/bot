package objectWorker

import (
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

func (o *ObjectWorkerT) InitWorker() {
	global.ServerState.SetObjectReady()
	go o.flow()
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

		threadList := [][]v1alpha1.TransferRequestT{}
		requestList := []v1alpha1.TransferRequestT{}
		currentThreads := 0
		requestIndex := 0
		requestsCount := 0
		for key, request := range transferRequestPool {
			requestList = append(requestList, request)
			global.TransferRequestPool.RemoveRequest(key)
			requestsCount++

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
		}

		logger.Logger.Infof("object worker status {requests: '%d', threads: '%d'}", requestsCount, currentThreads)
		wg.Wait()
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
