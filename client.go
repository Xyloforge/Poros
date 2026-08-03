package main

import (
	"log"
	"net"
	"os"
	"sync"
)

var cLog = log.New(os.Stdout, "[CLIENT]: ", log.Ldate|log.Ltime|log.Lshortfile)

func ConnectUDP() (*net.UDPConn, error) {
	rAddr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		cLog.Fatal("[WARN]: ", err)
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		cLog.Fatal("[WARN]: ", err)
		return nil, err
	}

	return conn, nil
}

var readBufferPool sync.Pool = sync.Pool{
	New: func() any {
		return new([1024]byte)
	},
}

func SendUDP(msg string, conn *net.UDPConn) {
	if len(msg) == 0 {
		return
	}

	var serviceType byte
	switch msg[0] {
	case '0':
		serviceType = byte(Create) // 0
	case '1':
		serviceType = byte(Join) // 1
	case '2':
		serviceType = byte(Leave) // 2
	case '3':
		serviceType = byte(Broad) // 3
	case '4':
		serviceType = byte(Stun) // 4
	case '5':
		serviceType = byte(Discover) // 5
	default:
		serviceType = byte(Unknown) // 99
	}
	payload := append([]byte{serviceType}, []byte(msg[1:])...)
	_, err := conn.Write(payload)
	if err != nil {
		cLog.Println("[WARN]: ", err)
	}
}

func StartClientListener(conn *net.UDPConn) {
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFrom(buffer)
			if err != nil {
				return
			}
			cLog.Printf("Server broadcast/response: %s\n", string(buffer[:n]))
		}
	}()
}
