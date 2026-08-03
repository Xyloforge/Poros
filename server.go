package main

import (
	"log"
	"net"
	"os"
)

func StartUDPServer() {
	logger := log.New(os.Stdout, "[SERVER]: ", log.Ldate|log.Ltime|log.Lshortfile)
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		logger.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Fatal(err)
	}

	mm := NewMapManager(conn)

	// Start HTTP dashboard for monitoring rooms & connected clients
	go StartHTTPServer(":8081", mm)

	logger.Println("UDP Server started")

	go func() {
		defer conn.Close()
		buffer := make([]byte, 1024)
		for {
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				logger.Println("[WARN] Read error:", err)
				continue
			}

			if n < 1 {
				continue // Ignore empty datagrams to prevent panics
			}

			msg := buffer[:n]

			serviceType := DetermineServiceType(buffer[0])

			// logger.Printf("[INFO] %s (%s): %s", remoteAddr.AddrPort(), serviceType.String(), msg)

			switch serviceType {
			case Create:
				mm.CreateRoom(remoteAddr)
			case Join:
				mm.JoinRoom(remoteAddr, roomKey(msg[1:]))
			case Leave:
				mm.LeaveRoom(remoteAddr)
			case Broad:
				mm.BroadCast(remoteAddr, msg[1:])
			default:
				_, err = conn.WriteToUDP([]byte("Unknown"), remoteAddr)
				if err != nil {
					logger.Println("[WARN] Write error:", err)
				}
			}

			// response := []byte("Message received!")
			// _, err = conn.WriteToUDP(response, remoteAddr)
			// if err != nil {
			// 	logger.Println("[WARN] Write error:", err)
			// }
		}

	}()
}
