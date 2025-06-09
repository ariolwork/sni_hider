package net_extentions

import (
	"fmt"
	"net"
)

var okMessage []byte = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")

func SetOK(s net.Conn) error {
	n, err := s.Write(okMessage)
	if err != nil || n == 0 {
		return fmt.Errorf("Failed to send ok status %s", err)
	}
	return nil
}
