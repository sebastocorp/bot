package databaseWorker

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/database"
)

type DatabaseWorkerT struct {
	Config v1alpha1.DatabaseWorkerConfigT

	DatabaseManager database.ManagerT
}

func (d *DatabaseWorkerT) InitWorker() {
	global.ServerState.SetDatabaseReady()
	go d.flow()
}

func (d *DatabaseWorkerT) flow() {
	for {
		// CONSUME REQUESTS TO MIGRATE FROM MAP OR WAIT
		databaseRequestPool := global.DatabaseRequestPool.GetPool()

		poolLen := len(databaseRequestPool)
		if poolLen == 0 {
			logger.Logger.Infof("database worker status {empty pool}")
			time.Sleep(2 * time.Second)
			continue
		}

		threadList := [][]v1alpha1.DatabaseRequestT{}
		requestList := []v1alpha1.DatabaseRequestT{}
		currentThreads := 0
		requestIndex := 0
		requestsCount := 0
		for key, request := range databaseRequestPool {
			requestList = append(requestList, request)
			global.DatabaseRequestPool.RemoveRequest(key)
			requestsCount++

			if requestIndex++; requestIndex >= global.Config.DatabaseWorker.RequestsByChildThread {
				threadList = append(threadList, requestList)
				requestIndex = 0
				requestList = []v1alpha1.DatabaseRequestT{}

				if currentThreads++; currentThreads >= global.Config.DatabaseWorker.MaxChildTheads {
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

			go d.processRequestList(&wg, requests)
		}

		logger.Logger.Infof("database worker status {requests: %d, threads: '%d'}", requestsCount, currentThreads)
		wg.Wait()
	}
}

func (d *DatabaseWorkerT) processRequestList(wg *sync.WaitGroup, requests []v1alpha1.DatabaseRequestT) {
	defer wg.Done()
	reqsStr := ""
	for _, req := range requests {
		reqsStr += req.String()
	}

	idStr := fmt.Sprintf("%x", md5.Sum([]byte(reqsStr)))

	logger.Logger.Infof("process database request list with id '%s', requests '%s'", idStr, reqsStr)
	err := d.DatabaseManager.InsertObjectListIfNotExist(requests)
	if err != nil {
		logger.Logger.Errorf("unable to process database request list with id '%s': %s", idStr, err.Error())
	} else {
		logger.Logger.Infof("success in process database request list with id '%s'", idStr)
	}
}
