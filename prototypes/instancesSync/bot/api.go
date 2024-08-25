package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIT struct {
	ctx        context.Context
	Port       string
	HttpServer *http.Server
}

// request/response types

type resInfoBodyT struct {
	Server apiServerT `json:"server"`
}

type apiServerT struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// API REST Functions

func (b *BotT) checkAPI(requestURL string) (err error) {
	res, err := http.Get(requestURL)
	if err != nil {
		log.Printf("api '%s' not ready: %s", requestURL, err.Error())
		return err
	}

	if res != nil && res.StatusCode == http.StatusOK {
		log.Printf("api '%s' is ready", requestURL)
	}

	return err
}

// curl -X GET http://headless-test.default.svc.cluster.local/instancename
func (b *BotT) getInstanceReady(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ready": "true"})
}

func (b *BotT) getInstanceInfo(c *gin.Context) {
	tmp := resInfoBodyT{Server: apiServerT(b.Server)}
	c.JSON(http.StatusOK, gin.H{"server": tmp.Server})
}

// func (b *BotT) deleteServer(c *gin.Context) {
// 	// bind request body with data struct
// 	var reqServer apiServerT
// 	if err := c.ShouldBindJSON(&reqServer); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// update servers pool
// 	ServerInstancesPool.RemoveServer(ServerT(reqServer))

// 	c.JSON(http.StatusOK, gin.H{"message": "server removed", "data": reqServer})
// }

func (b *BotT) InitAPI() {
	log.Print("Init API")

	router := gin.Default()
	router.GET("/instance/ready", b.getInstanceReady)
	router.GET("/instance/info", b.getInstanceInfo)
	// router.DELETE("/server", b.deleteServer)

	addr := fmt.Sprintf("0.0.0.0:%s", b.API.Port)

	b.API.ctx = context.Background()
	b.API.HttpServer = &http.Server{
		Addr:    addr,
		Handler: router.Handler(),
	}

	go func() {
		// service connections
		if err := b.API.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	// router.Run(addr)
}
