# Poros 🌀

**A high-performance NAT-traversal, UDP matchmaking, and relay engine written in Go from scratch.**

> *Poros* (Greek: πόρος — "the passage, the way through") connects peers across separate home networks behind NAT barriers via UDP hole-punching and relay fallback.

---

## 🚀 Features

- **UDP Datagram Core**: Fast, zero-allocation network transport built on Go's `net.UDPConn`.
- **Concurrency-Safe Room Engine**: Goroutine-managed room directory (`mapManager`) supporting room creation, room joining, leave mechanics, and intra-room broadcasting.
- **Real-Time Google Material Dashboard**: Embedded HTTP web UI at `http://localhost:8081` with live metrics, connected peer monitoring, and room occupancy status.
- **Binary Wire Protocol**: Compact 1-byte opcode headers for maximum efficiency over UDP.
- **Multi-Language Clients**: Includes both native Go (`client.go`) and Python (`client.py`) test clients with asynchronous background broadcast listeners.

---

## 🔌 Wire Protocol

Poros uses a compact binary header byte prepended to all UDP payloads:

| Opcode | Service | Header Byte | Payload Format | Example |
| :--- | :--- | :--- | :--- | :--- |
| **0** | `Create` | `0x00` | None | `0` (creates a 6-character room key) |
| **1** | `Join` | `0x01` | `<RoomKey>` | `1aBcDeF` (joins room `aBcDeF`) |
| **2** | `Leave` | `0x02` | None | `2` (leaves current room) |
| **3** | `Broad` | `0x03` | `<Message>` | `3Hello room!` (broadcasts to peers) |

---

## 🛠️ Quick Start

### 1. Start the Server

Run the UDP engine and web dashboard:

```bash
go run .
```

* **UDP Transport Listener**: `:8080`
* **Web Monitoring Dashboard**: `http://localhost:8081`

---

### 2. Connect via Python Client

Run the interactive Python test client in a new terminal window:

```bash
python3 client.py
```

#### Example Usage:
```text
Message to send: 0
[CLIENT]: Server response: room created 2G53Q7

Message to send: 3Hello everyone!
```

---

### 3. Connect via Go Client

Run the Go client:

```bash
go run client.go main.go util.go service.go
```

---

## 📊 Web Dashboard

Open **[http://localhost:8081](http://localhost:8081)** in your browser to view live server metrics:

- **Active Rooms Counter**: Total rooms currently managed by `mapManager`.
- **Connected Clients Counter**: Total active UDP peer endpoints.
- **Live Room Cards**: Displays room keys, occupancy progress bars (`X / 4 Clients`), and active client `IP:Port` addresses with instant auto-refresh.

---

## 📂 Project Structure

```text
poros/
├── server.go         # UDP server listener & opcode dispatcher
├── map.go            # Thread-safe Room & Client mapManager engine
├── service.go        # ServiceType binary byte enum definitions
├── web.go            # HTTP server & Google Material 3 web dashboard
├── util.go           # Fast zero-allocation room key generator
├── client.go         # Go UDP test client implementation
├── client.py         # Asynchronous Python test client
├── main.go           # Application entry point
├── server-roadmap.md # STUN / TURN / ICE implementation roadmap
└── README.md         # Project documentation
```

---

## 📜 Roadmap & Future Milestones

See [`server-roadmap.md`](./server-roadmap.md) for full architectural design and upcoming milestones (STUN public address discovery, UDP hole-punching, TURN relay, and ICE fallback).
