package main

type ServiceType byte

const (
	Create   ServiceType = iota // 0
	Join                        // 1
	Leave                       // 2
	Broad                       // 3
	Stun                        // 4
	Discover                    // 5
)

const Unknown ServiceType = 99

func (s ServiceType) String() string {
	switch s {
	case Create:
		return "Create"
	case Join:
		return "Join"
	case Broad:
		return "Broad"
	case Leave:
		return "Leave"
	case Stun:
		return "Stun"
	case Discover:
		return "Discover"
	default:
		return "Unknown"
	}
}

func DetermineServiceType(d byte) ServiceType {
	firstByte := ServiceType(d)
	switch firstByte {
	case Create, Join, Broad, Leave, Stun, Discover:
		return firstByte
	default:
		return Unknown
	}
}
