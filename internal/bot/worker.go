package bot

import (
	"bot/internal/logger"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WORKER Functions

func (b *BotT) executeTransferRequest(request TransferT) (err error) {
	// check if destination object already exist
	destExist, _, err := b.ObjectManager.S3ObjectExist(request.To)
	if err != nil {
		return err
	}

	if destExist {
		err = fmt.Errorf("object '%s' already exist in '%s' destination bucket", request.To.ObjectPath, request.To.BucketName)
		return err
	}

	// check if source object already exist
	sourceExist, sourceInfo, err := b.ObjectManager.GCSObjectExist(request.From)
	if err != nil {
		return err
	}

	if !sourceExist {
		err = fmt.Errorf("object '%s' NOT found in '%s' source bucket", request.From.ObjectPath, request.From.BucketName)
		return err
	}

	if len(sourceInfo.MD5) == 0 {
		err = fmt.Errorf("unable to transfer object '%s' without md5 assosiated in '%s' source bucket", request.From.ObjectPath, request.From.BucketName)
		return err
	}

	request.From.Etag = hex.EncodeToString(sourceInfo.MD5)
	request.To.Etag = request.From.Etag

	_, err = b.ObjectManager.TransferObjectFromGCSToS3(request.From, request.To)
	if err != nil {
		return err
	}

	// Get the object from the database
	_, occurrences, err := b.DatabaseManager.GetObject(request.To)
	if err != nil {
		return err
	}

	// Insert the object in the database
	if occurrences == 0 {
		err = b.DatabaseManager.InsertObject(request.To)
		if err != nil {
			return err
		}
	}

	return err
}

func (b *BotT) moveTransferRequest(serverName string, request TransferT) (err error) {
	pool := ServerInstancesPool.GetPool()
	serverToSend := ServerT{}
	for _, server := range pool {
		if server.Name == serverName {
			serverToSend = server
			break
		}
	}

	body := apiTransferRequestT{
		Transfer: request,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	http.DefaultClient.Timeout = 200 * time.Millisecond
	requestURL := fmt.Sprintf("http://%s:%s/transfer", serverToSend.Address, b.API.Port)
	respBody, err := http.Post(requestURL, headerContentTypeAppJson, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	defer respBody.Body.Close()

	return err
}

func (b *BotT) processTransferRequest(itemPath string, request TransferT) {
	logger.Logger.Infof("process '%s' transfer request '%v' in '%s'", itemPath, request, b.Server.Name)
	serverName := HashRing.GetNode(itemPath)

	if serverName != b.Server.Name {
		// send transfer request to owner
		logger.Logger.Infof("moving '%s' transfer request from '%s' to '%s'", itemPath, b.Server.Name, serverName)

		err := b.moveTransferRequest(serverName, request)
		if err != nil {
			logger.Logger.Errorf("unable to move '%s' transfer request to '%s'", itemPath, serverName)
		}

		return
	}

	logger.Logger.Infof("execute '%s' transfer request in '%s'", itemPath, b.Server.Name)
	err := b.executeTransferRequest(request)
	if err != nil {
		logger.Logger.Errorf("unable to execute '%s' transfer request in '%s': %s", itemPath, b.Server.Name, err.Error())
	} else {
		logger.Logger.Infof("success executing '%s' transfer request in '%s'", itemPath, b.Server.Name)
	}
}

func (b *BotT) workerFlow() {
	for {
		// CONSUME OBJECT TO MIGRATE FROM MAP OR WAIT
		transferRequestMap := TransferRequestPool.GetTransferRequestMap()

		for itemPath, request := range transferRequestMap {
			go b.processTransferRequest(itemPath, request)
			// remove request from pool
			TransferRequestPool.RemoveTransferRequest(itemPath)
		}

		time.Sleep(1 * time.Second) // REMOVE THIS IN THE END
	}
}

func (b *BotT) InitWorker() {
	go b.workerFlow()
}
