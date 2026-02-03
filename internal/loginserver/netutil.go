package loginserver

import (
	"net"
	"strings"
)

func remoteIP(conn net.Conn) string {
	if conn == nil {
		return ""
	}
	if tcp, ok := conn.RemoteAddr().(*net.TCPAddr); ok && tcp.IP != nil {
		return tcp.IP.String()
	}
	addr := conn.RemoteAddr().String()
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	if idx := strings.LastIndex(addr, ":"); idx > 0 {
		return addr[:idx]
	}
	return addr
}
