package objectWorker

import (
	"context"
	"sync"
	"time"

	"bot/api/v1alpha3"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/objectStorage"
	"bot/internal/pools"
)

type ObjectWorkerT struct {
	ctx    context.Context
	config *v1alpha3.BOTConfigT
	log    logger.LoggerT

	// ObjectManager objectStorage.ManagerT
	// hashring            *hashring.HashRingT
	objectRequestPool   *pools.ObjectRequestPoolT
	databaseRequestPool *pools.DatabaseRequestPoolT
	// serverInstancePool  *pools.ServerInstancesPoolT

	sources map[string]objectStorage.ObjectManagerI
}

// WORKER Functions

func NewObjectWorker(config *v1alpha3.BOTConfigT, objectPool *pools.ObjectRequestPoolT, dbPool *pools.DatabaseRequestPoolT) (ow *ObjectWorkerT, err error) {
	ow = &ObjectWorkerT{
		ctx:                 context.Background(),
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

	ow.sources = map[string]objectStorage.ObjectManagerI{}
	for _, sv := range config.ObjectWorker.Sources {
		ow.sources[sv.Name], err = objectStorage.GetManager(ow.ctx, sv)
		if err != nil {
			return ow, err
		}
	}

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

	emptyPoolLog := true
	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestPool := ow.objectRequestPool.GetPool()

		poolLen := len(transferRequestPool)
		if poolLen == 0 {
			if emptyPoolLog {
				ow.log.Debug("object request pool empty", logExtraFields)
				emptyPoolLog = false
			}
			time.Sleep(2 * time.Second)
			continue
		}
		emptyPoolLog = true

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
		back, backSource, err := ow.getBackendObject(request.Object)
		if err != nil {
			logExtraFields[global.LogFieldKeyExtraError] = err.Error()
			ow.log.Error("unable to get backend object route", logExtraFields)
			continue
		}
		front, frontSource, err := ow.getFrontendObject(request.Object)
		if err != nil {
			logExtraFields[global.LogFieldKeyExtraError] = err.Error()
			ow.log.Error("unable to get frontend object route", logExtraFields)
			continue
		}
		logExtraFields[global.LogFieldKeyExtraError] = global.LogFieldValueDefault
		logExtraFields[global.LogFieldKeyExtraObject] = front.String()
		logExtraFields[global.LogFieldKeyExtraBackendObject] = back.String()
		ow.log.Info("process object transfer request", logExtraFields)

		backobj, err := ow.sources[backSource].GetObject(back)
		if err != nil {
			logExtraFields[global.LogFieldKeyExtraError] = err.Error()
			ow.log.Error("unable to get backend object", logExtraFields)
			continue
		}
		defer backobj.Close()

		err = ow.sources[frontSource].PutObject(front, backobj)
		if err != nil {
			logExtraFields[global.LogFieldKeyExtraError] = err.Error()
			ow.log.Error("unable to put frontend object", logExtraFields)
			continue
		}

		ow.databaseRequestPool.AddRequest(pools.DatabaseRequestT{
			BucketName: front.Bucket,
			ObjectPath: front.Path,
			MD5:        backobj.GetMD5String(),
		})

		ow.log.Info("success in process object transfer request", logExtraFields)
	}
}
