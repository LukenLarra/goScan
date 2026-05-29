package scan

import (
	"net"
	"strconv"
	"strings"
	"time"
)

type MockDialer struct {
	OpenPorts map[int]bool
}

func (m MockDialer) Dial(network, address string, timeout time.Duration) (net.Conn, error) {
	parts := strings.Split(address, ":")
	portStr := parts[len(parts)-1]
	port, _ := strconv.Atoi(portStr)
	if m.OpenPorts[port] {
		client, server := net.Pipe()
		server.Close()
		return client, nil
	}
	return nil, &net.OpError{Op: "dial", Net: network, Addr: nil, Err: &net.DNSError{Err: "connection refused", Name: address, IsTemporary: false, IsTimeout: false}}
}
