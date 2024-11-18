package apiService

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bot/api/v1alpha3"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/pools"
)

type APIServiceT struct {
	config *v1alpha3.BOTConfigT
	log    logger.LoggerT

	ctx               context.Context
	objectRequestPool *pools.ObjectRequestPoolT
	httpServer        *http.Server
}

// API REST Functions

func NewApiService(config *v1alpha3.BOTConfigT, objectPool *pools.ObjectRequestPoolT) (a *APIServiceT) {
	a = &APIServiceT{
		config:            config,
		objectRequestPool: objectPool,
	}

	logCommon := global.GetLogCommonFields()
	logCommon[global.LogFieldKeyCommonInstance] = a.config.Name
	logCommon[global.LogFieldKeyCommonComponent] = global.LogFieldValueComponentAPIService
	a.log = logger.NewLogger(context.Background(), logger.GetLevel(a.config.APIService.LogLevel),
		logCommon,
	)

	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc(global.EndpointHealthz, a.getHealthz)
	mux.HandleFunc(global.EndpointInfo, a.getInfo)
	mux.HandleFunc(global.EndpointRequestTransfer, a.postTransferRequest)
	mux.HandleFunc(global.EndpointRequestObject, a.postTransferRequest)

	a.ctx = context.Background()
	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", a.config.APIService.Address, a.config.APIService.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return a
}

func (a *APIServiceT) Run() {
	logExtraFields := global.GetLogExtraFieldsAPI()

	global.ServerState.SetAPIReady()
	go func() {
		// service connections
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logExtraFields[global.LogFieldKeyExtraError] = err.Error()
			a.log.Fatal("unable to serve api", logExtraFields)
		}
	}()
}

func (a *APIServiceT) Shutdown() {
	logExtraFields := global.GetLogExtraFieldsAPI()

	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	if err := a.httpServer.Shutdown(ctx); err != nil {
		logExtraFields[global.LogFieldKeyExtraError] = err.Error()
		a.log.Fatal("error in service shutdown", logExtraFields)
	}

	cancel()
}

func (a *APIServiceT) getHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	result := "Unavailable"
	statusCode := http.StatusServiceUnavailable

	if global.ServerState.IsReady() {
		statusCode = http.StatusOK
		result = "OK"
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(result))
}

func (a *APIServiceT) getInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	server := pools.ServerT{
		Name:    a.config.Name,
		Address: a.config.APIService.Address,
	}

	w.Header().Set(global.HeaderContentType, global.HeaderContentTypeAppJson)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(server)
}

// example:
// curl -X POST
// http://bot-host/transfer --header "Content-Type: application/json"
// --data
// {
// 	"bucket":"backend-bucket",
// 	"path":"path/to/object"
// },

func (a *APIServiceT) postTransferRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	logExtraFields := global.GetLogExtraFieldsAPI()

	objectRequest := pools.ObjectRequestT{}
	if err := json.NewDecoder(r.Body).Decode(&objectRequest.Object); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)

		logExtraFields[global.LogFieldKeyExtraError] = err.Error()
		a.log.Error("object request decode error", logExtraFields)
		return
	}

	a.objectRequestPool.AddRequest(objectRequest)

	w.Header().Set(global.HeaderContentType, global.HeaderContentTypeAppJson)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(objectRequest.Object)

	logExtraFields[global.LogFieldKeyExtraObject] = objectRequest.Object.String()
	a.log.Info("object request added in pool", logExtraFields)
}
