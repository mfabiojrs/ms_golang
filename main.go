package main

import (
	"log"
)

func main() {
	serverManager := NewServerManager()
	tcpServer := NewTCPServer(serverManager)
	httpServer := NewHTTPServer(serverManager)

	log.Println("Starting HTTP server")
	go httpServer.Run("",8090)
	
	log.Println("Starting TCP server")
	go tcpServer.Run("",3333)

	log.Println("Starting ServerManager")
	serverManager.Run("",3333)
}