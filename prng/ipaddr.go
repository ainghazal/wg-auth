package prng

import (
	"fmt"
	"net"
)

// GetNthIP returns the nth IP address in a given CIDR range.
func GetNthIP(cidr string, n int) (string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR range")
	}

	// Convert IP to 4-byte slice for IPv4 addresses
	ip := ipNet.IP.To4()
	if ip == nil {
		return "", fmt.Errorf("only IPv4 addresses are supported")
	}

	// Convert IP to integer
	ipInt := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])

	// Get the network and broadcast addresses
	ones, bits := ipNet.Mask.Size()
	numAddresses := 1 << (bits - ones)
	if n < 0 || n >= numAddresses {
		return "", fmt.Errorf("n is out of the range for the given CIDR")
	}

	// Add n to the IP integer
	ipInt += uint32(n)

	// Convert integer back to IP
	newIP := net.IPv4(byte(ipInt>>24), byte(ipInt>>16), byte(ipInt>>8), byte(ipInt)).String()
	return newIP, nil
}
