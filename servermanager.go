package main

import (
	"log"
	"time"
	"strings"
)

const StaleAge float64 = 1//65  // Age in minutes where the server gets killed after no activity


type ServerManager struct {
	servers map[string]*ServerRecord
	udpPinger *UDPPinger
}

func NewServerManager() *ServerManager {
	return &ServerManager{
		servers: make(map[string]*ServerRecord),
		udpPinger: NewUDPPinger(),
	}
}

func (serverManager *ServerManager) Run(host string, port int) {
	go serverManager.udpPinger.Run("", 3333)
	for {
		serverManager.UpdateServers()
		time.Sleep(30*time.Second)
	}
}

// Attempts to register a server and returns the status on the 
func (serverManager *ServerManager) RegisterServer(server *ServerInfo) (*ServerRecord, string) {
	record := NewServerRecord(server)
	serverManager.servers[record.GetID()] = record

	initSuccess := serverManager.udpPinger.InitPing(record)

	if initSuccess {
		record.Status = RegistrationStatusSuccess
		go serverManager.udpPinger.PingLoop(record)
		log.Printf("Server %s successfully registered", record.GetID())
		return record, "success"
	} else {
		message := "failed to ping server"
		serverManager.removeServer(record)
		log.Printf("Failed to register server %s: %s", record.GetID(), message)
		return record, message
	}
}

func (serverManager *ServerManager) GetServerListString() string {
	commands := make([]string, 0, len(serverManager.servers))

	for _, record := range serverManager.servers {
		if record.Status == RegistrationStatusSuccess {
			commands = append(commands, record.AddServerString())
		}
	}

	return strings.Join(commands,"\n")
}

func (serverManager *ServerManager) UpdateServers() {
	for _, record := range serverManager.servers {
		if record.Status == RegistrationStatusSuccess {
			if record.TimeSincePong() > StaleAge {
				log.Printf("Unregistering %s (unresponsive)", record.GetID())
				serverManager.removeServer(record)
			}
		}
	}
}

func (serverManager *ServerManager) removeServer(record *ServerRecord) {
	record.Status = RegistrationStatusInvalid
	delete(serverManager.servers, record.GetID())
}
