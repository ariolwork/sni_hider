package handlers

import "net"

type Handler interface {
	Handle(c net.Conn) error
}
