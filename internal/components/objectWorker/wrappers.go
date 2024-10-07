package objectWorker

import (
	"strings"

	"bot/internal/managers/objectStorage"
	"bot/internal/pools"
)

func (ow *ObjectWorkerT) executeTransferRequest(request pools.ObjectRequestT, srcObject objectStorage.ObjectT) (err error) {
	sourceInfo, err := ow.ObjectManager.TransferObjectFromGCSToS3(srcObject, request.Object)
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

func (ow *ObjectWorkerT) getBackendObject(object objectStorage.ObjectT) (backend objectStorage.ObjectT) {
	backend = object
	switch ow.config.ObjectWorker.ObjectModification.Type {
	case "bucket":
		{
			if mods, ok := ow.config.ObjectWorker.ObjectModification.Mods[backend.Bucket]; ok {
				backend.Bucket = mods.Bucket
				backend.Path = strings.TrimPrefix(backend.Path, mods.RemovePrefix)
				backend.Path = mods.AddPrefix + backend.Path
			}
		}
	}

	return backend
}

// func (ow *ObjectWorkerT) moveTransferRequest(serverName string, request pools.ObjectRequestT) (err error) {
// 	pool := ow.serverInstancePool.GetPool()
// 	serverToSend := pools.ServerT{}
// 	for _, server := range pool {
// 		if server.Name == serverName {
// 			serverToSend = server
// 			break
// 		}
// 	}

// 	bodyBytes, err := json.Marshal(request)
// 	if err != nil {
// 		return err
// 	}

// 	http.DefaultClient.Timeout = 100 * time.Millisecond
// 	// TODO: add api port configuration to use here
// 	requestURL := fmt.Sprintf("http://%s:%s/transfer", serverToSend.Address, "8080")
// 	respBody, err := http.Post(requestURL, global.HeaderContentTypeAppJson, bytes.NewBuffer(bodyBytes))
// 	if err != nil {
// 		return err
// 	}
// 	defer respBody.Body.Close()

// 	return err
// }
