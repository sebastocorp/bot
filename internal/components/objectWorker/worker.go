package objectWorker

import (
	"context"
	"sync"
	"time"

	"bot/api/v1alpha2"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/objectStorage"
	"bot/internal/pools"
)

type ObjectWorkerT struct {
	config *v1alpha2.BOTConfigT
	log    logger.LoggerT

	ObjectManager objectStorage.ManagerT
	// hashring            *hashring.HashRingT
	objectRequestPool   *pools.ObjectRequestPoolT
	databaseRequestPool *pools.DatabaseRequestPoolT
	// serverInstancePool  *pools.ServerInstancesPoolT
}

// WORKER Functions

func NewObjectWorker(config *v1alpha2.BOTConfigT, objectPool *pools.ObjectRequestPoolT, dbPool *pools.DatabaseRequestPoolT) (ow *ObjectWorkerT, err error) {
	ow = &ObjectWorkerT{
		config:              config,
		objectRequestPool:   objectPool,
		databaseRequestPool: dbPool,
	}

	logCommon := global.GetLogCommonFields()
	logCommon[global.LogFieldKeyCommonInstance] = ow.config.Name
	logCommon[global.LogFieldKeyCommonComponent] = global.LogFieldValueComponentObjectWorker
	ow.log = logger.NewLogger(context.Background(),
		logger.GetLevel(ow.config.ObjectWorker.LogLevel),
		logCommon,
	)

	ow.ObjectManager, err = objectStorage.NewManager(
		context.Background(),
		ow.config.ObjectWorker.ObjectStorage.S3,
		ow.config.ObjectWorker.ObjectStorage.GCS,
	)

	return ow, err
}

func (ow *ObjectWorkerT) Run() {
	global.ServerState.SetObjectReady()
	go ow.flow()
}

func (ow *ObjectWorkerT) Shutdown() {
	// logExtraFields := maps.Clone(global.LogExtraFields)

	// transferRequestPool := ow.objectRequestPool.GetPool()
	// for key, request := range transferRequestPool {
	// 	ow.objectRequestPool.RemoveRequest(key)

	// 	logExtraFields[global.LogFieldKeyExtraCurrentRequest] = request.String()
	// 	ow.log.Debug("process object transfer request", logExtraFields)

	// 	// if global.Config.HashRingWorker.Enabled {
	// 	// 	serverName := global.HashRing.GetNode(request.To.ObjectPath)

	// 	// 	if serverName != global.Config.Name {
	// 	// 		// send transfer request to owner
	// 	// 		logger.Logger.Infof("moving object transfer request '%s' to '%s'", request.String(), serverName)

	// 	// 		err := ow.moveTransferRequest(serverName, request)
	// 	// 		if err == nil {
	// 	// 			return
	// 	// 		}

	// 	// 		logger.Logger.Errorf("unable to move object transfer request '%s' to '%s': %s", request.String(), serverName, err.Error())
	// 	// 	}
	// 	// }

	// 	err := ow.executeTransferRequest(request)
	// 	if err != nil {
	// 		logExtraFields[global.LogFieldKeyExtraError] = err.Error()
	// 		ow.log.Error("unable to process object transfer request", logExtraFields)
	// 	} else {
	// 		ow.log.Debug("success in process object transfer request", logExtraFields)
	// 	}
	// }
}

func (ow *ObjectWorkerT) flow() {
	logExtraFields := global.GetLogExtraFieldsObjectWorker()

	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestPool := ow.objectRequestPool.GetPool()

		poolLen := len(transferRequestPool)
		if poolLen == 0 {
			ow.log.Debug("object request pool empty", logExtraFields)
			time.Sleep(2 * time.Second)
			continue
		}

		threadList := [][]pools.ObjectRequestT{}
		requestList := []pools.ObjectRequestT{}
		currentThreads := 0
		requestIndex := 0
		requestsCount := 0
		for key, request := range transferRequestPool {
			requestList = append(requestList, request)
			ow.objectRequestPool.RemoveRequest(key)
			requestsCount++

			if requestIndex++; requestIndex >= ow.config.ObjectWorker.RequestsByChildThread {
				threadList = append(threadList, requestList)
				requestIndex = 0
				requestList = []pools.ObjectRequestT{}

				if currentThreads++; currentThreads >= ow.config.ObjectWorker.MaxChildTheads {
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

		logExtraFields[global.LogFieldKeyExtraActiveRequestCount] = requestsCount
		logExtraFields[global.LogFieldKeyExtraActiveThreadCount] = currentThreads
		logExtraFields[global.LogFieldKeyExtraCurrentPoolLength] = poolLen - requestsCount
		ow.log.Debug("object worker handle requests", logExtraFields)

		wg.Wait()
	}
}

func (ow *ObjectWorkerT) processRequestList(wg *sync.WaitGroup, requests []pools.ObjectRequestT) {
	defer wg.Done()

	logExtraFields := global.GetLogExtraFieldsObjectWorker()

	for _, request := range requests {
		backend := ow.getBackendObject(request.Object)
		logExtraFields[global.LogFieldKeyExtraObject] = request.Object.String()
		logExtraFields[global.LogFieldKeyExtraBackendObject] = backend.String()
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

		err := ow.executeTransferRequest(request, backend)
		if err != nil {
			logExtraFields[global.LogFieldKeyExtraError] = err.Error()
			ow.log.Error("unable to process object transfer request", logExtraFields)
		} else {
			ow.log.Debug("success in process object transfer request", logExtraFields)
		}
	}
}
