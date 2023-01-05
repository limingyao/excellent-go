package net

import (
	"errors"
	"net"
)

// This code is borrowed from https://github.com/uber/tchannel-go/blob/dev/localip.go

// scoreAddr scores how likely the given addr is to be a remote address and returns the
// IP to use when listening. Any address which receives a negative score should not be used.
// Scores are calculated as:
// -1 for any unknown IP addresses.
// +300 for IPv4 addresses
// +100 for non-local addresses, extra +100 for "up" interfaces.
// +100 for routable addresses
// -50 for local mac addr.
func scoreAddr(iface net.Interface, addr net.Addr) (int, net.IP) {
	var ip net.IP
	if netAddr, ok := addr.(*net.IPNet); ok {
		ip = netAddr.IP
	} else if netIP, ok := addr.(*net.IPAddr); ok {
		ip = netIP.IP
	} else {
		return -1, nil
	}

	var score int
	if ip.To4() != nil {
		score += 300
	}
	if iface.Flags&net.FlagLoopback == 0 && !ip.IsLoopback() {
		score += 100
		if iface.Flags&net.FlagUp != 0 {
			score += 100
		}
	}
	if _, routable := isRoutableIP("ip", ip); routable {
		score -= 25
	}
	if isLocalMacAddr(iface.HardwareAddr) {
		score -= 50
	}
	return score, ip
}

// filter is an interface filter which returns false if the interface is _not_ to listen on
func listenAddr(interfaces []net.Interface, filter func(iface net.Interface) bool) (net.HardwareAddr, net.IP, error) {
	bestScore := -1
	var bestIP net.IP
	var bestMac net.HardwareAddr
	// Select the highest scoring IP as the best IP.
	for _, iface := range interfaces {
		if !filter(iface) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			// Skip this interface if there is an error.
			continue
		}

		for _, addr := range addrs {
			score, ip := scoreAddr(iface, addr)
			if score > bestScore {
				bestScore = score
				bestIP = ip
				bestMac = iface.HardwareAddr
			}
		}
	}

	if bestScore == -1 {
		return nil, nil, errors.New("no addresses to listen on")
	}

	return bestMac, bestIP, nil
}

func ListenAddr(filters ...func(iface net.Interface) bool) (net.HardwareAddr, net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, err
	}
	return listenAddr(interfaces, func(iface net.Interface) bool {
		for _, filter := range filters {
			if filter != nil && !filter(iface) {
				return false
			}
		}
		return true
	})
}

// If the first octet's second least-significant-bit is set, then it's local.
// https://en.wikipedia.org/wiki/MAC_address#Universal_vs._local
func isLocalMacAddr(addr net.HardwareAddr) bool {
	if len(addr) == 0 {
		return false
	}
	return addr[0]&2 == 2
}

func isRoutableIP(network string, ip net.IP) (net.IP, bool) {
	if !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsGlobalUnicast() {
		return nil, false
	}
	switch network {
	case "ip4":
		if ip := ip.To4(); ip != nil {
			return ip, true
		}
	case "ip6":
		if ip.IsLoopback() { // addressing scope of the loopback address depends on each implementation
			return nil, false
		}
		if ip := ip.To16(); ip != nil && ip.To4() == nil {
			return ip, true
		}
	default:
		if ip := ip.To4(); ip != nil {
			return ip, true
		}
		if ip := ip.To16(); ip != nil {
			return ip, true
		}
	}
	return nil, false
}
