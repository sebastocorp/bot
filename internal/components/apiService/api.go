package apiService

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/pools"

	"github.com/gin-gonic/gin"
)

type APIServiceT struct {
	ctx    context.Context
	config v1alpha1.APIServiceConfigT
	log    logger.LoggerT

	objectRequestPool *pools.ObjectRequestPoolT
	httpServer        *http.Server
}

// API REST Functions

func NewApiService(config v1alpha1.APIServiceConfigT) (a *APIServiceT) {
	a = &APIServiceT{
		config: config,
	}

	level, err := logger.GetLevel(a.config.LogLevel)
	if err != nil {
		log.Fatalf("unable to get api service loglevel: %s", err.Error())
	}

	a.log = logger.NewLogger(context.Background(), level, map[string]any{
		"service":   "bot",
		"component": "api",
	})

	router := gin.Default()
	router.GET(global.EndpointHealth, a.getHealth)
	router.GET(global.EndpointInfo, a.getInfo)
	router.POST(global.EndpointRequestObject, a.postTransferRequest)

	a.ctx = context.Background()
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", a.config.Address, a.config.Port),
		Handler: router.Handler(),
	}

	return a
}

func (a *APIServiceT) Run() {

	global.ServerState.SetAPIReady()
	go func() {
		// service connections
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatal("unable to serve api", map[string]any{
				"error": err.Error(),
			})
		}
	}()
}

func (a *APIServiceT) Shutdown() {

	logExtraFields := map[string]any{
		"error": "none",
	}

	ctx, cancel := context.WithTimeout(a.ctx, 1*time.Second)
	if err := a.httpServer.Shutdown(ctx); err != nil {
		logExtraFields["error"] = err.Error()
		a.log.Fatal("error in shutdown: %s", logExtraFields)
	}

	cancel()
}

func (a *APIServiceT) getHealth(c *gin.Context) {
	if !global.ServerState.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (a *APIServiceT) getInfo(c *gin.Context) {
	server := v1alpha1.ServerT{
		// Name:    a.config.Name,
		Address: a.config.Address,
	}
	c.JSON(http.StatusOK, server)
}

// example:
// curl -X POST
// http://bot-host/transfer --header "Content-Type: application/json"
// --data
// {
// 	"transfer":{
// 		"from":{
// 			"bucket":"backend-bucket",
// 			"path":"path/to/object"
// 		},
// 		"to":{
// 			"bucket":"frontend-bucket",
// 			"path":"path/to/object"
// 		}
// 	}
// }

func (a *APIServiceT) postTransferRequest(c *gin.Context) {
	logExtraFields := map[string]any{
		"error": "none",
	}

	transfer := v1alpha1.TransferRequestT{}
	if err := c.ShouldBindJSON(&transfer.To); err != nil {
		logExtraFields["error"] = err.Error()
		a.log.Error("error parsing transfer request", logExtraFields)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a.objectRequestPool.AddRequest(transfer)
	c.JSON(http.StatusOK, gin.H{"message": "OK"})

	a.log.Info("transfer request added in pool", logExtraFields)
}
