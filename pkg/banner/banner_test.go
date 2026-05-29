package banner

import (
	"net"
	"testing"
	"time"
)

func TestIdentify(t *testing.T) {
	tests := []struct {
		banner  []byte
		service string
	}{
		{[]byte("SSH-2.0-OpenSSH_8.9"), "SSH"},
		{[]byte("HTTP/1.1 200 OK"), "HTTP"},
		{[]byte("220 ftp.example.com FTP server ready"), "FTP/SMTP"},
		{[]byte("+OK POP3 ready"), "POP3"},
		{[]byte("\x16\x03\x01\x00\xf1"), "TLS/SSL"},
		{[]byte("UNKNOWN PROTOCOL"), ""},
	}

	for _, tt := range tests {
		got := Identify(tt.banner)
		if got != tt.service {
			t.Errorf("Identify(%q) = %q, want %q", tt.banner, got, tt.service)
		}
	}
}

func TestGrabWithNetPipe(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	go func() {
		defer server.Close()
		server.Write([]byte("SSH-2.0-OpenSSH_8.9\r\n"))
	}()

	data := Grab(client, time.Second)
	if Identify(data) != "SSH" {
		t.Errorf("esperaba SSH, got %q", Identify(data))
	}
}
