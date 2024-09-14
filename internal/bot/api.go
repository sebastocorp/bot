package bot

import (
	"bot/internal/logger"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	headerContentTypeAppJson = "application/json"
)

type APIT struct {
	ctx        context.Context
	Port       string
	HttpServer *http.Server
}

// request/response types

type apiInfoRequestT struct {
	Server ServerT `json:"server"`
}

type apiTransferRequestT struct {
	Transfer TransferT `json:"transfer"`
}

// API REST Functions

func (b *BotT) getInstanceStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func (b *BotT) getInstanceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"server": b.Server})
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

func (b *BotT) postTransfer(c *gin.Context) {
	transfer := apiTransferRequestT{}
	if err := c.ShouldBindJSON(&transfer); err != nil {
		logger.Logger.Errorf("error parsing transfer request: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	TransferRequestPool.AddTransferRequest(TransferT{
		From: transfer.Transfer.From,
		To:   transfer.Transfer.To,
	})

	logger.Logger.Infof("transfer request '%v' added in pool", transfer)

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (b *BotT) InitAPI() {
	router := gin.Default()
	router.GET("/status", b.getInstanceStatus)
	router.GET("/info", b.getInstanceInfo)
	router.POST("/transfer", b.postTransfer)

	addr := fmt.Sprintf("%s:%s", b.Server.Address, b.API.Port)

	b.API.ctx = context.Background()
	b.API.HttpServer = &http.Server{
		Addr:    addr,
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := b.API.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatalf("error closing API in '%s': %s\n", b.Server.Name, err.Error())
		}
	}()
}
