# Poros

**A NAT-traversal and matchmaking server in Go — built from scratch.**

Poros connects two game clients on different networks. Peers behind separate home
routers can't reach each other directly; Poros is the public middleman that lets
them find each other by a room code and then talk — directly via UDP hole-punching
where possible, relayed through the server where not.

The name is Greek: *poros* (πόρος) — "the passage, the way through." Its opposite,
*aporia*, means impasse. That's the job: a way through the NAT barrier.

> This repo is a from-scratch implementation — STUN, TURN, and ICE-style logic
> written by hand rather than pulled from a library. The roadmap below is the
> build order, with what to verify at each stage.

---

## Background — why this is needed

Two ideas explain the entire design.

**NAT blocks unsolicited inbound traffic.** Every device on a home network has a
private IP (`192.168.x.x`); the whole network shares one public IP from the ISP.
The router (via Network Address Translation) rewrites outbound packets to the
public address and remembers the mapping, so replies find their way back. But an
*unsolicited* inbound packet — a peer on another network trying to initiate — has
no matching table entry and gets dropped. So two peers behind separate NATs cannot
directly connect. This is the problem Poros exists to solve.

**Outbound always works.** Both peers can freely connect *out* to a machine with a
public IP. So the strategy is never to ask a peer to accept an inbound connection.
Instead both peers reach out to a common public server, which either introduces
them (so they can connect directly) or relays between them.

**Transport is UDP.** Real-time games favor UDP over TCP — a lost packet shouldn't
stall everything behind it waiting on a resend. Hole-punching is also far more
practical over UDP. Poros is UDP end to end.

### The building blocks (implemented by hand here)
- **STUN** — lets a peer discover its own public IP:port as seen from outside.
- **TURN** — relays traffic between peers that can't connect directly.
- **ICE** — the overall strategy: gather addresses, try direct, fall back to relay.
- **Hole punching** — the technique for a direct UDP link through two NATs.

---

## Roadmap

Each milestone is independently runnable and testable, in build order.

### M0 — UDP foundation
**Goal:** move raw bytes over UDP in Go.
**Scope:** a listener that receives and prints datagrams; a client that sends one
and receives a reply.
**Verify:** message and reply round-trip on localhost.
**Understanding check:** how does this differ from a TCP listener, and why is there
no connection to "accept"?

### M1 — Two-client relay
**Goal:** the server forwards messages between two clients that never talk directly.
**Scope:** the server tracks two client addresses and forwards each one's messages
to the other.
**Verify:** three processes on localhost (server + two clients); a message typed in
one appears in the other. This is a working relay — the fallback path of the final
system.
**Understanding check:** where does the "who are the two clients" state live, and
what breaks when a third client connects?

### M2 — Rooms and matchmaking
**Goal:** many clients self-organize into rooms by code.
**Scope:** a client hosts a room and receives a generated code; another joins by
code; the server maps code → room members and relays only within a room.
**Key engineering:**
- The room table is shared across many goroutines — make it concurrency-safe
  (mutex-guarded map, or a single owner goroutine receiving commands over a
  channel; both are worth implementing to compare).
- Rooms must not leak — track last-seen time per client and reap dead rooms from a
  background goroutine.
**Verify:** open a room, join from two clients, confirm isolation between rooms;
kill a client and confirm cleanup.
**Understanding check:** which concurrency model guards the room table, and what's
the rule for declaring a client gone?

### M3 — Deploy the relay (cross-network)
**Goal:** prove it works between genuinely different networks, not localhost.
**Scope:** run the server on a small VPS with a public IP and an open UDP port
(covers firewall / security-group config).
**Verify:** one client on home wifi, another on a different network (e.g. mobile
data), connected via room code. At this stage all traffic flows through the VPS —
pure relay, which always works. Direct connection comes next.
**Understanding check:** trace one message's full path — origin, through NAT, to
VPS, through the other NAT, to the peer — naming each hop.

### M4 — STUN-lite: public-address discovery
**Goal:** a peer learns its own public IP:port as seen from outside — the
prerequisite for a direct connection.
**Scope:** a server endpoint that replies to a client with the source address the
client's packet arrived from.
**Verify:** compare the address the server sees against the address the client
believes it has; on a home network they differ, and that difference is the NAT
mapping.
**Understanding check:** why can't a client determine its own public address
without asking an external server?

### M5 — UDP hole punching
**Goal:** a direct peer-to-peer link; the server drops out of the data path.
**Mechanism:** with both peers in a room, the server (knowing both public addresses
from M4) sends each peer the other's address. Both peers then send UDP packets
toward each other simultaneously. Each outbound packet creates a NAT table entry
expecting a reply from the other's address, so the previously-dropped packets now
pass — the peers have punched holes in both NATs and can talk directly.
**Scope:** server-side coordination (hand both peers each other's address, timed
together) and client-side punch logic (send until a reply arrives, then switch to
the direct path).
**Verify:** two peers on different networks connect directly; server logs confirm
it's no longer forwarding.
**NAT types to know:** full-cone, address-restricted, port-restricted, and
symmetric. Punching works against most, but symmetric NAT assigns a different
public port per destination, invalidating the discovered address and defeating the
punch. This is a known limitation of the technique — and the reason the relay
still exists.
**Understanding check:** describe the punch as a timeline of packets and NAT table
entries, and explain why symmetric NAT breaks it.

### M6 — ICE-lite: direct with relay fallback
**Goal:** always-connect behavior.
**Scope:** attempt the hole-punch; if a direct link is established within a
timeout, use it; otherwise fall back to relaying through the server. The client
runs either mode transparently to the game layer above.
**Verify:** exercise both paths — two home networks typically yield a direct link;
a peer on mobile data can trigger the relay fallback. The game connects either way.
**Understanding check:** where does the direct-vs-relay switch happen, and what's
the punch timeout?

### M7 — Hardening
**Goal:** survive real-world conditions.
**Scope (priority order):**
- Reconnection — a dropped peer rejoins its room.
- Clean disconnect — notify the room so it doesn't wait on a ghost.
- Keepalives — NAT mappings expire on inactivity; peers send periodic packets to
  keep the hole open.
- Abuse guards — room-code collisions, capacity limits, rate limiting.
- (Later) transport encryption.
**Understanding check:** enumerate every way a connection can die and Poros's
response to each.

---

## Client integration

Poros is transport only; it's agnostic to the game engine on top. The single
integration surface is the **join protocol** — the wire format for the messages
between client and server: hello, host-room, join-by-code, and the server's
responses. Define that message format after M2 (before M3) and document it in
[`PROTOCOL.md`](./PROTOCOL.md) so any client can implement against it independently.

---

## References

- **STUN / TURN / ICE** — the standards Poros reimplements in miniature.
  RFCs: STUN (5389/8489), TURN (8656), ICE (8445). Plain-English WebRTC connection
  explainers cover the same introduce → punch → relay flow.
- **UDP hole punching** — Ford, Srisuresh & Kegel, *Peer-to-Peer Communication
  Across Network Address Translators*.
- **Go networking** — the standard library `net` package (UDP connections,
  datagram read/write).
- **NAT behavior** — full-cone / restricted / port-restricted / symmetric NAT
  semantics, which explain M5's success and failure cases.

---

## Status

Work in progress — see the roadmap above for current milestone.