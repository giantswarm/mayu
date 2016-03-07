package pxemgr

import "net"

func incIP(ip net.IP) net.IP {
	ip = ip.To4()
	numIP := uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
	numIP++
	newIP := net.IPv4(byte(numIP>>24&0xff), byte(numIP>>16&0xff), byte(numIP>>8&0xff), byte(numIP&0xff))

	if newIP.IsMulticast() {
		return incIP(newIP)
	}
	return newIP
}

// ip less or equal
func ipLessThanOrEqual(ip net.IP, upperBound net.IP) bool {
	if ip[3] < upperBound[3] {
		return true
	}
	if ip[2] < upperBound[2] {
		return true
	}
	if ip[1] < upperBound[1] {
		return true
	}
	if ip[0] <= upperBound[0] {
		return true
	}

	return false
}
