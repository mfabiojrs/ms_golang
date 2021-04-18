package main

import (
	"fmt"
	"net"
	"time"
)

type RegistrationStatus int
const (
	RegistrationStatusPending RegistrationStatus = iota // Server is in process of registering
	RegistrationStatusSuccess // Server has succeeded the registration process
	RegistrationStatusInvalid // Server has become unresponsive
)

type ServerInfo struct {
	IP net.IP
	Port int
	Name string
	Version int
	Key int
}

func NewServerInfo(ip net.IP, port int, name string, version int, key int) *ServerInfo {
	return &ServerInfo {
		IP: ip,
		Port: port,
		Name: name,
		Version: version,
		Key: key,
	}
}

type ServerRecord struct {
	Info *ServerInfo
	Status RegistrationStatus
	LastPong time.Time
}

func NewServerRecord(serverInfo *ServerInfo) *ServerRecord {
	return &ServerRecord {
		Info: serverInfo,
		Status: RegistrationStatusPending,
		LastPong: time.Now(),
	}
}

func (sr *ServerRecord) TimeSincePong() float64 {
	return time.Since(sr.LastPong).Minutes()
}

func (sr *ServerRecord) AddServerString() string {
	return fmt.Sprintf("addserver %s %d", sr.Info.IP.String(), sr.Info.Port)
}

func (sr *ServerRecord) GetID() string {
	return fmt.Sprintf("%s:%d", sr.Info.IP.String(), sr.Info.Port)
}