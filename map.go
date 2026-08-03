package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type roomKey string
type clientKey string

type Room struct {
	RoomKey   roomKey
	MaxClient int
	Clients   map[clientKey]*Client
}

type Client struct {
	UDPAddr *net.UDPAddr
}

type RoomInfo struct {
	RoomKey   string   `json:"roomKey"`
	MaxClient int      `json:"maxClient"`
	Clients   []string `json:"clients"`
}

type MapState struct {
	TotalRooms   int        `json:"totalRooms"`
	TotalClients int        `json:"totalClients"`
	Rooms        []RoomInfo `json:"rooms"`
}

type MapManager interface {
	CreateRoom(clientAddr *net.UDPAddr)
	JoinRoom(clientAddr *net.UDPAddr, rk roomKey)
	LeaveRoom(clientAddr *net.UDPAddr)
	BroadCast(clientAddr *net.UDPAddr, message []byte)
	Stun(clientAddr *net.UDPAddr) string
	Discover(clientAddr *net.UDPAddr)
	GetState() MapState
}

var _ MapManager = (*mapManager)(nil)

type mapManager struct {
	logger    *log.Logger
	conn      *net.UDPConn
	ClientMap map[clientKey]*Room
	RoomMap   map[roomKey]*Room

	createRoomQ chan *net.UDPAddr
	joinRoomQ   chan joinRoomData
	leaveRoomQ  chan *net.UDPAddr
	broadCastQ  chan broadCastData
	discoverQ   chan *net.UDPAddr
	getStateQ   chan chan MapState
}

func NewMapManager(conn *net.UDPConn) MapManager {
	logger := log.New(os.Stdout, "[MAP MANAGER]: ", log.Ldate|log.Ltime|log.Lshortfile)
	m := mapManager{
		logger:    logger,
		conn:      conn,
		ClientMap: map[clientKey]*Room{},
		RoomMap:   map[roomKey]*Room{},

		createRoomQ: make(chan *net.UDPAddr, 100),
		joinRoomQ:   make(chan joinRoomData, 100),
		leaveRoomQ:  make(chan *net.UDPAddr, 100),
		getStateQ:   make(chan chan MapState, 100),
		broadCastQ:  make(chan broadCastData, 100),
	}

	go m.startWorker()
	return &m
}

type broadCastData struct {
	clientAddr *net.UDPAddr
	message    []byte
}

type joinRoomData struct {
	clientAddr *net.UDPAddr
	rk         roomKey
}

func genRoomKey() roomKey {
	return roomKey(RandString6())
}

func getClientKey(clientAddr *net.UDPAddr) clientKey {
	return clientKey(clientAddr.String())
}

func (m *mapManager) startWorker() {
	for {
		select {
		case broadCastData := <-m.broadCastQ:
			m.broadCast(broadCastData.clientAddr, broadCastData.message)
		case clientAddr := <-m.discoverQ:
			m.discover(clientAddr)
		case clientAddr := <-m.createRoomQ:
			rKey := genRoomKey()
			if _, ok := m.RoomMap[rKey]; ok {
				res := fmt.Sprintf("room key already exist: %s", rKey)
				m.logger.Printf("[WARN] %s\n", res)
				m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
				continue
			}

			cKey := getClientKey(clientAddr)
			newClient := Client{
				UDPAddr: clientAddr,
			}

			newRoom := Room{
				RoomKey:   rKey,
				MaxClient: 4,
				Clients: map[clientKey]*Client{
					cKey: &newClient,
				},
			}

			m.RoomMap[rKey] = &newRoom
			m.ClientMap[cKey] = &newRoom

			res := fmt.Sprintf("room created %s", rKey)
			m.logger.Printf("[INFO] %s\n", res)
			m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())

		case clientAddr := <-m.leaveRoomQ:
			cKey := getClientKey(clientAddr)
			if _, ok := m.ClientMap[cKey]; !ok {
				res := fmt.Sprintf("client not in any room %s", cKey)
				m.logger.Printf("[INFO] %s\n", res)
				m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
				continue
			}

			r := m.ClientMap[cKey]
			rKey := r.RoomKey

			delete(m.ClientMap, cKey)
			delete(r.Clients, cKey)
			res := fmt.Sprintf("client %s left room %s", cKey, rKey)
			m.logger.Printf("[INFO] %s\n", res)
			m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())

			if len(r.Clients) == 0 {
				delete(m.RoomMap, rKey)
				res = fmt.Sprintf("room %s close with no more client", rKey)
				m.logger.Printf("[INFO] %s\n", res)
				m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())

			}

		case joinRoomData := <-m.joinRoomQ:
			rKey := joinRoomData.rk
			if _, ok := m.RoomMap[rKey]; !ok {
				res := fmt.Sprintf("room not exist: %s", rKey)
				m.logger.Printf("[WARN] %s\n", res)
				m.conn.WriteToUDPAddrPort([]byte(res), joinRoomData.clientAddr.AddrPort())
				continue
			}

			cKey := getClientKey(joinRoomData.clientAddr)
			if _, ok := m.ClientMap[cKey]; ok {
				res := fmt.Sprintf("client %s already in room: %s", cKey, rKey)
				m.logger.Printf("[INFO] %s\n", res)
				m.conn.WriteToUDPAddrPort([]byte(res), joinRoomData.clientAddr.AddrPort())
				continue
			}

			r := m.RoomMap[rKey]
			if len(r.Clients) >= r.MaxClient {
				res := fmt.Sprintf("room already full: %s", rKey)
				m.logger.Printf("[INFO] %s\n", res)
				m.conn.WriteToUDPAddrPort([]byte(res), joinRoomData.clientAddr.AddrPort())
				continue
			}

			c := Client{
				UDPAddr: joinRoomData.clientAddr,
			}
			r.Clients[cKey] = &c

			m.ClientMap[cKey] = r
			res := fmt.Sprintf("client %s joined room %s, room cap %d/%d", cKey, rKey, len(r.Clients), r.MaxClient)
			m.logger.Printf("[INFO] %s\n", res)
			m.conn.WriteToUDPAddrPort([]byte(res), joinRoomData.clientAddr.AddrPort())

		case respChan := <-m.getStateQ:
			rooms := make([]RoomInfo, 0, len(m.RoomMap))
			for rKey, room := range m.RoomMap {
				clientList := make([]string, 0, len(room.Clients))
				for cKey := range room.Clients {
					clientList = append(clientList, string(cKey))
				}
				rooms = append(rooms, RoomInfo{
					RoomKey:   string(rKey),
					MaxClient: room.MaxClient,
					Clients:   clientList,
				})
			}
			respChan <- MapState{
				TotalRooms:   len(m.RoomMap),
				TotalClients: len(m.ClientMap),
				Rooms:        rooms,
			}
		}
	}
}

func (m *mapManager) CreateRoom(clientAddr *net.UDPAddr) {
	select {
	case m.createRoomQ <- clientAddr:
	default:
		res := "server is busy try again later"
		m.logger.Printf("[WARN] CreateRoom %s\n", res)
		m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
	}
}

func (m *mapManager) JoinRoom(clientAddr *net.UDPAddr, rk roomKey) {
	data := joinRoomData{
		clientAddr: clientAddr,
		rk:         rk,
	}
	select {
	case m.joinRoomQ <- data:
	default:
		res := "server is busy try again later"
		m.logger.Printf("[WARN] JoinRoom %s\n", res)
		m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
	}
}

func (m *mapManager) LeaveRoom(clientAddr *net.UDPAddr) {
	select {
	case m.leaveRoomQ <- clientAddr:
	default:
		res := "server is busy try again later"
		m.logger.Printf("[WARN] CreateRoom %s\n", res)
		m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
	}
}

func (m *mapManager) BroadCast(clientAddr *net.UDPAddr, message []byte) {
	data := broadCastData{
		clientAddr: clientAddr,
		message:    message,
	}

	select {
	case m.broadCastQ <- data:
	default:
		res := "server is busy try again later"
		m.logger.Printf("[WARN] BroadCast %s\n", res)
		m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
	}
}

func (m *mapManager) broadCast(clientAddr *net.UDPAddr, message []byte) {
	cKey := getClientKey(clientAddr)

	if _, ok := m.ClientMap[cKey]; !ok {
		res := fmt.Sprintf("client %s is not in any room", cKey)
		m.logger.Printf("[WARN] BroadCast %s\n", res)
		m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
		return
	}

	m.conn.WriteToUDPAddrPort([]byte("ok"), clientAddr.AddrPort())
	r := m.ClientMap[cKey]

	for key, client := range r.Clients {
		if key == cKey {
			continue
		}

		_, err := m.conn.WriteToUDPAddrPort(message, client.UDPAddr.AddrPort())
		if err != nil {
			m.logger.Printf("[WARN] failed broadcasting %s to %s\n", string(message), client.UDPAddr.AddrPort())
		}
	}
}

func (m *mapManager) GetState() MapState {
	respChan := make(chan MapState, 1)
	select {
	case m.getStateQ <- respChan:
		return <-respChan
	case <-time.After(2 * time.Second):
		return MapState{}
	}
}

func (m mapManager) Stun(clientAddr *net.UDPAddr) string {
	res := clientAddr.AddrPort().String()
	m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
	return res
}

func (m mapManager) Discover(clientAddr *net.UDPAddr) {
	select {
	case m.discoverQ <- clientAddr:
	default:
		res := "server is busy try again later"
		m.logger.Printf("[WARN] Discover %s\n", res)
		m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
	}
}

func (m mapManager) discover(clientAddr *net.UDPAddr) {
	cKey := getClientKey(clientAddr)
	if _, ok := m.ClientMap[cKey]; !ok {
		res := fmt.Sprintf("client %s is not in any room", cKey)
		m.logger.Printf("[WARN] Discover %s\n", res)
		m.conn.WriteToUDPAddrPort([]byte(res), clientAddr.AddrPort())
		return
	}

	r := m.ClientMap[cKey]

	result := ""
	for _, c := range r.Clients {
		if getClientKey(c.UDPAddr) == cKey {
			continue
		}

		result = fmt.Sprintf("%s,%s", result, string(getClientKey(clientAddr)))
	}

	m.conn.WriteToUDPAddrPort([]byte(result), clientAddr.AddrPort())
}
