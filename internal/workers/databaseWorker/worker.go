package databaseWorker

import (
	"sync"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/managers/database"
)

type DatabaseWorkerT struct {
	Config v1alpha1.DatabaseWorkerConfigT

	DatabaseManager database.ManagerT
}

func (d *DatabaseWorkerT) processRequest(wg *sync.WaitGroup, requests []v1alpha1.DatabaseRequestT) {
	defer wg.Done()

	logger.Logger.Infof("process database requests '%v'", requests)

	objects := []v1alpha1.DatabaseRequestT{}
	for _, req := range requests {
		objects = append(objects, v1alpha1.DatabaseRequestT{
			BucketName: req.BucketName,
			ObjectPath: req.ObjectPath,
			MD5:        req.MD5,
		})
	}
	// Get the object from the database
	err := d.DatabaseManager.InsertObjectsIfNotExist(objects)
	if err != nil {
		logger.Logger.Errorf("unable to process database requests '%v': %s", requests, err.Error())
	} else {
		logger.Logger.Infof("success processing database requests '%v'", requests)
	}
}

func (d *DatabaseWorkerT) flow() {
	for {
		// CONSUME REQUESTS TO MIGRATE FROM MAP OR WAIT
		databaseRequestPool := global.DatabaseRequestPool.GetPool()

		threadList := [][]v1alpha1.DatabaseRequestT{}
		requestList := []v1alpha1.DatabaseRequestT{}
		threadIndex := 0
		requestIndex := 0
		for _, request := range databaseRequestPool {
			requestList = append(requestList, request)

			if requestIndex++; requestIndex >= d.Config.InsertsByConnection {
				threadList = append(threadList, requestList)
				requestIndex = 0
				requestList = []v1alpha1.DatabaseRequestT{}

				if threadIndex++; threadIndex >= d.Config.ParallelRequests {
					break
				}
			}
		}

		if len(requestList) > 0 {
			threadList = append(threadList, requestList)
		}

		wg := sync.WaitGroup{}
		for _, list := range threadList {
			wg.Add(1)

			go d.processRequest(&wg, list)

			global.DatabaseRequestPool.RemoveRequests(list)
		}
		wg.Wait()
	}
}

func (d *DatabaseWorkerT) InitWorker() {
	global.ServerState.SetDatabaseReady()
	go d.flow()
}
