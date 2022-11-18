package testing

import (
	"net"
	"time"
)

// PortOpen returns true if there's a socket listening on
// the specified listenAddress.
func PortOpen(listenAddress string) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", listenAddress, timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		defer conn.Close()
		return true
	}

	return false
}
