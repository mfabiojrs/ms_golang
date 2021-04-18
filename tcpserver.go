package main

import (
	"fmt"
	"net"
	"log"
	"strings"
)

type Client struct {
	conn *net.TCPConn
	serverRecord *ServerRecord
}

func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		conn: conn,
	}
}

type TCPServer struct {
	serverManager *ServerManager
	running bool
	listener *net.TCPListener
}

func NewTCPServer(serverManager *ServerManager) *TCPServer{
	return &TCPServer {
		serverManager: serverManager,
		running: false,
	}
}

func (tcpserver *TCPServer) Stop() error {
	tcpserver.listener.Close()
	return nil
}

func (tcpserver *TCPServer) Run(host string, port int) error {
	if tcpserver.running {
		return nil
	}
	var err error
	laddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d",host,port))
	if err != nil {
		return err
	}

	tcpserver.listener, err = net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}
	log.Printf("Listening on %s:%d", host, port)

	for {
		conn, err := tcpserver.listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
		}
		go tcpserver.handleRequest(conn)
	}
}

func (tcpserver *TCPServer) handleRequest(conn *net.TCPConn) {
	buf := make([]byte, 4096)
	client := NewClient(conn) // Create client struct to contain the state for this connection

	for {
		reqLen, err := conn.Read(buf)
		if err != nil {
			log.Println("Error reading:", err.Error())
			break
		}

		shouldClose, err := tcpserver.handleCommand(client, string(buf[:reqLen]))
		if err != nil {
			log.Println("Error handling command: ", err.Error())
			break
		}
		if shouldClose {
			break
		}
	}

	conn.Close()
}

func (tcpserver *TCPServer) handleCommand(client *Client, command string) (bool, error) {
	command = strings.TrimSpace(command)

	log.Printf("Parsing command from %s: %s", client.conn.RemoteAddr().String(), command)

	if strings.HasPrefix(command, "list ") ||  command == "list" {
		var name string
		var version int
		_, err := fmt.Sscanf(command, "list %s %d", &name, &version)
		if err != nil {
			return true, err
		} else {
			log.Printf("Sending serverlist to %s %s (%d)", client.conn.RemoteAddr().String(), name, version)
			client.conn.Write([]byte(tcpserver.serverManager.GetServerListString()))
			return true, nil
		}
	}
	if strings.HasPrefix(command, "regserv ") {
		var port int
		var name string
		var version int
		var key int
		n, err := fmt.Sscanf(command, "regserv %d %s %d %d", &port, &name, &version, &key)
		if err != nil {
			return true, err
		}
		if n < 1 {
			return true, fmt.Errorf("Invalid number of arguments to regserv")
		}

		if (port < 0 || port > 0xFFFF-1) {
			client.conn.Write([]byte("failreg invalid port\n"))
			return true, fmt.Errorf("Port is out of range")
		}
		if (client.serverRecord != nil && client.serverRecord.Info.Port != port) {
			client.conn.Write([]byte("failreg invalid port\n"))
			return true, fmt.Errorf("Port is getting reuseds")
		}

		remoteAddr := client.conn.RemoteAddr()
		tcpAddr, _ := remoteAddr.(*net.TCPAddr)

		log.Printf("Received server registration request from %s. Port %d Name: %s Version %d Key %d", tcpAddr.IP.String(), port, name, version, key)
		serverInfo := NewServerInfo(tcpAddr.IP, port, name, version, key)
		record, message := tcpserver.serverManager.RegisterServer(serverInfo)

		if record.Status == RegistrationStatusSuccess {
			client.conn.Write([]byte("succreg\n"))
		} else {
			client.conn.Write([]byte(fmt.Sprintf("failreg %s\n", message)))
		}
		
		return false, nil
	}

	return true, fmt.Errorf("No Valid command found")
}