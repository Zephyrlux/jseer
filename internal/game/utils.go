package game

import (
	"bytes"
	"net"
)

func writeIP(buf *bytes.Buffer, ip string) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		buf.Write([]byte{127, 0, 0, 1})
		return
	}
	ip4 := parsed.To4()
	if ip4 == nil {
		buf.Write([]byte{127, 0, 0, 1})
		return
	}
	buf.Write(ip4)
}
