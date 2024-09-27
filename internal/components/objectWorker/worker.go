package objectWorker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/hashring"
	"bot/internal/managers/objectStorage"
	"bot/internal/pools"
)

type ObjectWorkerT struct {
	config v1alpha1.ObjectWorkerConfigT
	log    logger.LoggerT

	ObjectManager       objectStorage.ManagerT
	hashring            *hashring.HashRingT
	objectRequestPool   *pools.ObjectRequestPoolT
	databaseRequestPool *pools.DatabaseRequestPoolT
	serverInstancePool  *pools.ServerInstancesPoolT
}

// WORKER Functions

func NewObjectWorker(config v1alpha1.ObjectWorkerConfigT) (ow *ObjectWorkerT, err error) {
	ow = &ObjectWorkerT{
		config: config,
	}

	if ow.config.MaxChildTheads <= 0 {
		err = fmt.Errorf("config option objectWorker.maxChildTheads with value '%d', must be a number > 0",
			ow.config.MaxChildTheads,
		)
		return ow, err
	}

	if ow.config.RequestsByChildThread <= 0 {
		err = fmt.Errorf("config option objectWorker.requestsByChildThread with value '%d', must be a number > 0",
			ow.config.MaxChildTheads,
		)
		return ow, err
	}

	ow.ObjectManager, err = objectStorage.NewManager(
		context.Background(),
		ow.config.ObjectStorage.S3,
		ow.config.ObjectStorage.GCS,
	)

	return ow, err
}

func (ow *ObjectWorkerT) InitWorker() {
	global.ServerState.SetObjectReady()
	go ow.flow()
}

func (ow *ObjectWorkerT) flow() {
	logExtraFields := map[string]any{
		"error":    "none",
		"requests": "none",
		"threads":  "none",
		"pool":     "none",
	}

	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestPool := ow.objectRequestPool.GetPool()

		poolLen := len(transferRequestPool)
		if poolLen == 0 {
			ow.log.Debug("object request pool empty", logExtraFields)
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
			ow.objectRequestPool.RemoveRequest(key)
			requestsCount++

			if requestIndex++; requestIndex >= ow.config.RequestsByChildThread {
				threadList = append(threadList, requestList)
				requestIndex = 0
				requestList = []v1alpha1.TransferRequestT{}

				if currentThreads++; currentThreads >= ow.config.MaxChildTheads {
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

			go ow.processRequestList(&wg, requests)
		}

		logExtraFields["requests"] = requestsCount
		logExtraFields["threads"] = currentThreads
		logExtraFields["pool"] = poolLen - requestsCount
		ow.log.Debug("object worker handle requests", logExtraFields)

		wg.Wait()
	}
}

func (ow *ObjectWorkerT) processRequestList(wg *sync.WaitGroup, requests []v1alpha1.TransferRequestT) {
	defer wg.Done()

	logExtraFields := map[string]any{
		"error":   "none",
		"request": "none",
	}

	for _, request := range requests {
		logExtraFields["request"] = request.String()
		ow.log.Debug("process object transfer request", logExtraFields)

		// if global.Config.HashRingWorker.Enabled {
		// 	serverName := global.HashRing.GetNode(request.To.ObjectPath)

		// 	if serverName != global.Config.Name {
		// 		// send transfer request to owner
		// 		logger.Logger.Infof("moving object transfer request '%s' to '%s'", request.String(), serverName)

		// 		err := ow.moveTransferRequest(serverName, request)
		// 		if err == nil {
		// 			return
		// 		}

		// 		logger.Logger.Errorf("unable to move object transfer request '%s' to '%s': %s", request.String(), serverName, err.Error())
		// 	}
		// }

		err := ow.executeTransferRequest(request)
		if err != nil {
			logExtraFields["error"] = err.Error()
			ow.log.Error("unable to process object transfer request", logExtraFields)
		} else {
			ow.log.Debug("success in process object transfer request", logExtraFields)
		}
	}
}
