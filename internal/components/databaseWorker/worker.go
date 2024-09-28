package databaseWorker

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/database"
	"bot/internal/pools"
)

type DatabaseWorkerT struct {
	config *v1alpha1.BOTConfigT
	log    logger.LoggerT

	databaseRequestPool *pools.DatabaseRequestPoolT
	databaseManager     database.ManagerT
}

func NewDatabaseWorker(config *v1alpha1.BOTConfigT) (dw *DatabaseWorkerT, err error) {
	dw = &DatabaseWorkerT{
		config: config,
	}

	level, err := logger.GetLevel(dw.config.DatabaseWorker.LogLevel)
	if err != nil {
		level = logger.INFO
	}

	dw.log = logger.NewLogger(context.Background(), level, map[string]any{
		"service":   "bot",
		"component": "databaseWorker",
	})

	dw.databaseManager, err = database.NewManager(context.Background(),
		dw.config.DatabaseWorker.Database,
	)

	return dw, err
}

func (d *DatabaseWorkerT) Run() {
	global.ServerState.SetDatabaseReady()
	go d.flow()
}

func (d *DatabaseWorkerT) Shutdown() {
}

func (dw *DatabaseWorkerT) flow() {
	logExtraFields := map[string]any{
		"error":    "none",
		"requests": "none",
		"threads":  "none",
		"pool":     "none",
	}

	for {
		// CONSUME REQUESTS TO MIGRATE FROM MAP OR WAIT
		databaseRequestPool := dw.databaseRequestPool.GetPool()

		poolLen := len(databaseRequestPool)
		if poolLen == 0 {
			dw.log.Debug("database request pool empty", logExtraFields)
			time.Sleep(2 * time.Second)
			continue
		}

		threadList := [][]pools.DatabaseRequestT{}
		requestList := []pools.DatabaseRequestT{}
		currentThreads := 0
		requestIndex := 0
		requestsCount := 0
		for key, request := range databaseRequestPool {
			requestList = append(requestList, request)
			dw.databaseRequestPool.RemoveRequest(key)
			requestsCount++

			if requestIndex++; requestIndex >= dw.config.DatabaseWorker.RequestsByChildThread {
				threadList = append(threadList, requestList)
				requestIndex = 0
				requestList = []pools.DatabaseRequestT{}

				if currentThreads++; currentThreads >= dw.config.DatabaseWorker.MaxChildTheads {
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

			go dw.processRequestList(&wg, requests)
		}

		logExtraFields["requests"] = requestsCount
		logExtraFields["threads"] = currentThreads
		logExtraFields["pool"] = poolLen - requestsCount
		dw.log.Debug("database worker handle requests", logExtraFields)
		wg.Wait()
	}
}

func (dw *DatabaseWorkerT) processRequestList(wg *sync.WaitGroup, requests []pools.DatabaseRequestT) {
	defer wg.Done()
	reqsStr := ""
	for _, req := range requests {
		reqsStr += req.String()
	}

	idStr := fmt.Sprintf("%x", md5.Sum([]byte(reqsStr)))

	logExtraFields := map[string]any{
		"error":        "none",
		"request_id":   idStr,
		"request_list": reqsStr,
	}

	dw.log.Info("process database request list", logExtraFields)
	err := dw.databaseManager.InsertObjectListIfNotExist(requests)
	if err != nil {
		logExtraFields["error"] = err.Error()
		dw.log.Error("unable to process database request list", logExtraFields)
	} else {
		dw.log.Info("success in process database request list", logExtraFields)
	}
}
