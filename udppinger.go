package main

import (
	"log"
	"net"
	"fmt"
	"time"
)

const PingInterval float64 = 1//12 // Time in minutes between ping attempts
const InitialPingMaxAttempts int = 5
const InitialPingInterval float64 = 12

type UDPPinger struct {
	channels map[string]chan int
	conn *net.UDPConn
}

func NewUDPPinger() *UDPPinger{
	return &UDPPinger{
		channels: make(map[string]chan int),
	}
}

func (udpPinger *UDPPinger) Run(host string, port int) {
	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d",host,port))
	udpPinger.conn, err = net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	p :=  make([]byte, 5000)
	for {
		_, addr, err := udpPinger.conn.ReadFromUDP(p)// TODO: Switch to ReadMsgUDP
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Recieved pong from:", addr)
		pongID := fmt.Sprintf("%s:%d",addr.IP.String(),addr.Port-1)
		if pongChan, ok := udpPinger.channels[pongID]; ok {
			pongChan <- 1
		} else {
			log.Printf("Unable to find pong channel for %s", pongID)
		}
	}
}

func (udpPinger *UDPPinger) sendQuery(serverRecord *ServerRecord) {
	log.Printf("Sending ping to %s",serverRecord.GetID())
	dst, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverRecord.Info.IP.String(), serverRecord.Info.Port+1))
	if err != nil {
		log.Println(err)
	}

	// The connection can write data to the desired address.
	_, err = udpPinger.conn.WriteTo([]byte{1}, dst)
	if err != nil {
		log.Println(err)
	}
}

func (udpPinger *UDPPinger) InitPing(serverRecord *ServerRecord) bool {
	pongChan := make(chan int)
	udpPinger.channels[serverRecord.GetID()] = pongChan

	initSuccess := false

	func() {
		for i := 0; i < InitialPingMaxAttempts; i++ {
			udpPinger.sendQuery(serverRecord)
			select {
			case <-pongChan:
				initSuccess = true
				serverRecord.LastPong = time.Now()
				log.Printf("Initial ping success for %s", serverRecord.GetID())
				return
			case <-time.After(time.Duration(InitialPingInterval) * time.Second):
				log.Printf("Initial ping timeout for %s, attempt %d", serverRecord.GetID(), i)
			}
		}
	}()

	if !initSuccess {
		delete(udpPinger.channels, serverRecord.GetID())
	}

	return initSuccess
}

func (udpPinger *UDPPinger) PingLoop(serverRecord *ServerRecord) {
	pongChan := udpPinger.channels[serverRecord.GetID()]

	go func() {
		defer close(pongChan)
		for {
			time.Sleep(time.Duration(PingInterval) * time.Minute)

			if (serverRecord.Status != RegistrationStatusSuccess) {
				return
			}
			udpPinger.sendQuery(serverRecord)
		}
	}()

	go func() {
		for range pongChan {
			serverRecord.LastPong = time.Now()
		}
	}()
}