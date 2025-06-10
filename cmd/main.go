package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"tcp_sni_splitter/internal/handlers"
)

func main() {
	l := log.New(os.Stdout, "dpi: ", 0)
	subscribeListener(l, 3020)
}

func subscribeListener(log *log.Logger, port int) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("Cannot subsribe on port %s", err)
	}
	defer l.Close()

	tcpHandler := handlers.NewTcpHandler(log)
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Cannot accept connrection %s", err)
		}
		go func(c net.Conn) {
			defer conn.Close()
			err = tcpHandler.Handle(c)
			if err != nil {
				log.Printf("handler exception: %s", err)
			}
		}(conn)
	}
}
