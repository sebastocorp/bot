package databaseWorker

import (
	"sync"

	"bot/api/v1alpha1"
	"bot/internal/database"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/pools"
)

type DatabaseWorkerT struct {
	Config v1alpha1.DatabaseWorkerConfigT
	Server pools.ServerT

	DatabaseManager database.ManagerT
}

func (d *DatabaseWorkerT) processRequest(wg *sync.WaitGroup, requests []pools.DatabaseRequestT) {
	defer wg.Done()

	logger.Logger.Infof("process database requests '%v'", requests)

	objects := []database.ObjectT{}
	for _, req := range requests {
		objects = append(objects, database.ObjectT{
			Bucket: req.BucketName,
			Path:   req.ObjectPath,
			MD5:    req.MD5,
		})
	}
	// Get the object from the database
	err := d.DatabaseManager.InsertObjectsIfNotExist(objects)
	if err != nil {
		logger.Logger.Errorf("unable to process database requests '%v'", requests)
	} else {
		logger.Logger.Infof("success processing database requests '%v'", requests)
	}
}

func (d *DatabaseWorkerT) workerFlow() {
	for {
		// CONSUME REQUESTS TO MIGRATE FROM MAP OR WAIT
		databaseRequestPool := global.DatabaseRequestPool.GetPool()

		requestsList := [][]pools.DatabaseRequestT{}

		for i := 0; i < d.Config.ParallelRequests; i++ {
			requestsList = append(requestsList, []pools.DatabaseRequestT{})
		}

		requestsListIndex := 0
		inserts := 0
		for _, request := range databaseRequestPool {
			requestsList[requestsListIndex] = append(requestsList[requestsListIndex], request)

			if inserts++; inserts >= d.Config.InsertsByConnection {
				inserts = 0

				if requestsListIndex++; requestsListIndex >= d.Config.ParallelRequests {
					break
				}
			}
		}

		wg := sync.WaitGroup{}
		for _, list := range requestsList {
			wg.Add(1)

			go d.processRequest(&wg, list)

			global.DatabaseRequestPool.RemoveRequests(list)
		}
		wg.Wait()
	}
}

func (d *DatabaseWorkerT) InitWorker() {
	go d.workerFlow()
}
