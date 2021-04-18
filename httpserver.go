package main


import (
	"log"
	"fmt"
	"github.com/gin-gonic/gin"
)

type ClientRequest struct {
	Action  string `form:"action"`
	Name string `form:"name"`
	Version int `form:"version"`
	Build int `form:"build"`
}

type HTTPServer struct {
	serverManager *ServerManager
}

func NewHTTPServer(serverManager *ServerManager) *HTTPServer {
	return &HTTPServer {
		serverManager: serverManager,
	}
}

func (httpServer *HTTPServer) Run(host string, port int) {
	route := gin.Default()
	route.GET("/retrieve.do", httpServer.retrieve)
	route.Run(fmt.Sprintf("%s:%d",host,port))
}

// ms.cubers.net/retrieve.do?action=list&name=Blooptoop&version=1202&build=2
func (httpServer *HTTPServer) retrieve(c *gin.Context) {
	var clientRequest ClientRequest
	if err := c.BindQuery(&clientRequest); err != nil {
		log.Printf("Client request error: %v", err)
		return
	}

	log.Printf("%s: %s - %s (%d-%d)", c.ClientIP(), clientRequest.Action, clientRequest.Name, clientRequest.Version, clientRequest.Build)
	c.String(200, httpServer.serverManager.GetServerListString())
}