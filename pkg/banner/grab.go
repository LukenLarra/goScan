package banner

import (
	"net"
	"time"
)

const (
	readSize  = 256
	httpProbe = "HEAD / HTTP/1.0\r\n\r\n"
)

func Grab(conn net.Conn, timeout time.Duration) []byte {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	buf := make([]byte, readSize)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(timeout))
		_, _ = conn.Write([]byte(httpProbe))

		_ = conn.SetReadDeadline(time.Now().Add(timeout))
		n, _ = conn.Read(buf)
	}

	return buf[:n]
}
