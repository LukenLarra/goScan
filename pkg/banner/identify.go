package banner

import "bytes"

var signatures = []struct {
	prefix  []byte
	service string
}{
	{[]byte("SSH-"), "SSH"},
	{[]byte("HTTP/"), "HTTP"},
	{[]byte("220 "), "FTP/SMTP"},
	{[]byte("220-"), "FTP"},
	{[]byte("+OK"), "POP3"},
	{[]byte("* OK"), "IMAP"},
	{[]byte("RFB "), "VNC"},
	{[]byte("\x16\x03"), "TLS/SSL"},
	{[]byte("SMTP"), "SMTP"},
}

func Identify(banner []byte) string {
	for _, sig := range signatures {
		if bytes.HasPrefix(banner, sig.prefix) {
			return sig.service
		}
	}
	return ""
}
