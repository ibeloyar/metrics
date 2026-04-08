package netlib

import "net"

// GetOutboundIP returns the preferred outbound IP address of the machine.
// It establishes a temporary UDP connection to 8.8.8.8:80 and returns
// the local address of that connection as a string.
// If the connection cannot be established, it returns an empty string.
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
