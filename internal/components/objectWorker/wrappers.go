package objectWorker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/pools"
)

func (ow *ObjectWorkerT) executeTransferRequest(request pools.ObjectRequestT) (err error) {
	// check if destination object already exist
	// destInfo, err := ow.ObjectManager.S3ObjectExist(request.Object)
	// if err != nil {
	// 	return err
	// }
	// request.Object.Info = destInfo

	backend, err := ow.getBackendObject(request.Object)
	if err != nil {
		return err
	}

	sourceInfo, err := ow.ObjectManager.TransferObjectFromGCSToS3(backend, request.Object)
	if err != nil {
		return err
	}

	request.Object.Info = sourceInfo

	ow.databaseRequestPool.AddRequest(pools.DatabaseRequestT{
		BucketName: request.Object.Bucket,
		ObjectPath: request.Object.Path,
		MD5:        request.Object.Info.MD5,
	})

	return err
}

func (ow *ObjectWorkerT) getBackendObject(object v1alpha1.ObjectT) (backend v1alpha1.ObjectT, err error) {

	if ow.config.ObjectWorker.Source.Type == "bucket" {
		if mods, ok := ow.config.ObjectWorker.Source.ObjectMods[object.Bucket]; ok {
			backend.Bucket = mods.Bucket
			backend.Path = object.Path
			backend.Path = strings.TrimPrefix(backend.Path, mods.RemovePrefix)
			backend.Path = mods.AddPrefix + backend.Path

			return backend, err
		}
	}

	err = fmt.Errorf("object modification not defined")

	return backend, err
}

func (ow *ObjectWorkerT) moveTransferRequest(serverName string, request pools.ObjectRequestT) (err error) {
	pool := ow.serverInstancePool.GetPool()
	serverToSend := pools.ServerT{}
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
