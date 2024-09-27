package objectWorker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/global"
)

func (ow *ObjectWorkerT) executeTransferRequest(request v1alpha1.TransferRequestT) (err error) {
	// check if destination object already exist
	destInfo, err := ow.ObjectManager.S3ObjectExist(request.To)
	if err != nil {
		return err
	}
	request.To.Info = destInfo

	if !destInfo.Exist {
		sourceInfo, err := ow.ObjectManager.TransferObjectFromGCSToS3(request.From, request.To)
		if err != nil {
			return err
		}

		request.To.Info = sourceInfo
	}

	ow.databaseRequestPool.AddRequest(v1alpha1.DatabaseRequestT{
		BucketName: request.To.BucketName,
		ObjectPath: request.To.ObjectPath,
		MD5:        request.To.Info.MD5,
	})

	return err
}

func (ow *ObjectWorkerT) moveTransferRequest(serverName string, request v1alpha1.TransferRequestT) (err error) {
	pool := ow.serverInstancePool.GetPool()
	serverToSend := v1alpha1.ServerT{}
	for _, server := range pool {
		if server.Name == serverName {
			serverToSend = server
			break
		}
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	http.DefaultClient.Timeout = 100 * time.Millisecond
	// TODO: add api port configuration to use here
	requestURL := fmt.Sprintf("http://%s:%s/transfer", serverToSend.Address, "8080")
	respBody, err := http.Post(requestURL, global.HeaderContentTypeAppJson, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	defer respBody.Body.Close()

	return err
}
