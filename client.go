package main

import (
	"log"
	"net"
	"os"
	"sync"
	"time"
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
	default:
		serviceType = byte(Unknown) // 99
	}
	payload := append([]byte{serviceType}, []byte(msg[1:])...)
	_, err := conn.Write(payload)
	if err != nil {
		cLog.Println("[WARN]: ", err)
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	bufPtr := readBufferPool.Get().(*[1024]byte)
	defer readBufferPool.Put(bufPtr)

	buffer := bufPtr[:]
	n, _, err := conn.ReadFrom(buffer)
	if err != nil {
		cLog.Println("[WARN]: ", err)
	}

	cLog.Printf("Server response: %s\n", string(buffer[:n]))
}
