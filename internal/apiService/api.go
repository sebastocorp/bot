package apiService

import (
	"context"
	"fmt"
	"net/http"

	"bot/internal/global"
	"bot/internal/logger"
	"bot/internal/pools"

	"github.com/gin-gonic/gin"
)

type APIServiceT struct {
	Ctx        context.Context
	HttpServer *http.Server
}

// API REST Functions

func (a *APIServiceT) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (a *APIServiceT) getInfo(c *gin.Context) {
	c.JSON(http.StatusOK, global.ServerReference)
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

func (a *APIServiceT) postTransfer(c *gin.Context) {
	transfer := pools.TransferT{}
	if err := c.ShouldBindJSON(&transfer); err != nil {
		logger.Logger.Errorf("error parsing transfer request: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	global.TransferRequestPool.AddTransferRequest(transfer)

	logger.Logger.Infof("transfer request '%v' added in pool", transfer)

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (a *APIServiceT) InitAPI() {
	router := gin.Default()
	router.GET("/health", a.getHealth)
	router.GET("/info", a.getInfo)
	router.POST("/transfer", a.postTransfer)

	addr := fmt.Sprintf("%s:%s", global.ServerConfig.APIService.Address, global.ServerConfig.APIService.Port)

	a.Ctx = context.Background()
	a.HttpServer = &http.Server{
		Addr:    addr,
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := a.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatalf("error closing '%s' API: %s\n", global.ServerConfig.Name, err.Error())
		}
	}()
}
