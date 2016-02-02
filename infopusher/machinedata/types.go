package machinedata

import "net"

type HostData struct {
	Serial       string
	NetDevs      []NetDev
	IPMIAddress  net.IP
	ConnectedNIC string
}

type NetDev struct {
	Name       string
	MacAddress string
}
