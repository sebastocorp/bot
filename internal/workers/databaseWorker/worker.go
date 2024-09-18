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

func (d *DatabaseWorkerT) processRequest(wg *sync.WaitGroup, request v1alpha1.DatabaseRequestT) {
	defer wg.Done()

	logger.Logger.Infof("process database request '%s'", request.String())
	err := d.DatabaseManager.InsertObjectIfNotExist(request)
	if err != nil {
		logger.Logger.Errorf("unable to process database request '%s': %s", request.String(), err.Error())
	} else {
		logger.Logger.Infof("success in process database request '%s'", request.String())
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

func (d *DatabaseWorkerT) flow() {
	for {
		// CONSUME REQUESTS TO MIGRATE FROM MAP OR WAIT
		databaseRequestPool := global.DatabaseRequestPool.GetPool()

		poolLen := len(databaseRequestPool)
		if poolLen == 0 {
			time.Sleep(2 * time.Second)
		}

		wg := sync.WaitGroup{}
		currentThreads := 0
		for dbKey, request := range databaseRequestPool {
			wg.Add(1)

			go d.processRequest(&wg, request)
			global.DatabaseRequestPool.RemoveRequest(dbKey)

			if currentThreads++; currentThreads >= global.Config.DatabaseWorker.MaxChildTheads {
				break
			}
		}

		logger.Logger.Infof("current database worker status {threads: '%d', pool_length: '%d'}", currentThreads, poolLen)
		wg.Wait()
	}
}

func (d *DatabaseWorkerT) multiRequestFlow() {
	for {
		// CONSUME REQUESTS TO MIGRATE FROM MAP OR WAIT
		databaseRequestPool := global.DatabaseRequestPool.GetPool()

		poolLen := len(databaseRequestPool)
		if poolLen == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		threadList := [][]v1alpha1.DatabaseRequestT{}
		requestList := []v1alpha1.DatabaseRequestT{}
		currentThreads := 0
		requestIndex := 0
		for _, request := range databaseRequestPool {
			requestList = append(requestList, request)

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
			global.DatabaseRequestPool.RemoveRequests(requests)
		}

		logger.Logger.Infof("current database worker status {threads: '%d', pool_length: '%d'}", currentThreads, poolLen)
		wg.Wait()
	}
}

func (d *DatabaseWorkerT) InitWorker() {
	global.ServerState.SetDatabaseReady()
	if global.Config.DatabaseWorker.RequestsByChildThread > 0 {
		go d.multiRequestFlow()
	} else {
		go d.flow()
	}
}
