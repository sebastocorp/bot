package apiService

import (
	"context"
	"fmt"
	"net/http"

	"bot/api/v1alpha1"
	"bot/internal/global"
	"bot/internal/logger"

	"github.com/gin-gonic/gin"
)

type APIServiceT struct {
	Ctx        context.Context
	HttpServer *http.Server
}

// API REST Functions

func (a *APIServiceT) getHealth(c *gin.Context) {
	if !global.ServerState.IsReady() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (a *APIServiceT) getInfo(c *gin.Context) {
	server := v1alpha1.ServerT{
		Name:    global.Config.Name,
		Address: global.Config.APIService.Address,
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
	transfer := v1alpha1.TransferRequestT{}
	if err := c.ShouldBindJSON(&transfer); err != nil {
		logger.Logger.Errorf("error parsing transfer request: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	global.TransferRequestPool.AddRequest(transfer)

	logger.Logger.Infof("transfer request '%v' added in pool", transfer)

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (a *APIServiceT) postDatabaseRequest(c *gin.Context) {
	request := v1alpha1.DatabaseRequestT{}
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Logger.Errorf("error parsing database request: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	global.DatabaseRequestPool.AddRequest(request)

	logger.Logger.Infof("database request '%v' added in pool", request)

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (a *APIServiceT) InitAPI() {
	router := gin.Default()
	router.GET(global.EndpointHealth, a.getHealth)
	router.GET(global.EndpointInfo, a.getInfo)
	router.POST(global.EndpointRequestObject, a.postTransferRequest)
	router.POST(global.EndpointRequestDatabase, a.postDatabaseRequest)

	addr := fmt.Sprintf("%s:%s", global.Config.APIService.Address, global.Config.APIService.Port)

	a.Ctx = context.Background()
	a.HttpServer = &http.Server{
		Addr:    addr,
		Handler: router.Handler(),
	}

	global.ServerState.SetAPIReady()
	go func() {
		// service connections
		if err := a.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatalf("error closing '%s' API: %s\n", global.Config.Name, err.Error())
		}
	}()
}
