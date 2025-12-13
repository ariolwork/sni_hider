package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"tcp_sni_splitter/internal/handlers"
	"time"
)

func main() {
	l := log.New(os.Stdout, "dpi: ", 0)
	go func() {
		printGoroutines(l)
	}()
	subscribeListener(l, 3020)
}

func subscribeListener(log *log.Logger, port int) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("Cannot subsribe on port %s", err)
	}
	defer l.Close()

	tcpHandler := handlers.NewTcpHandler(log, false)
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Cannot accept connrection %s", err)
		}
		go func(c net.Conn) {
			defer c.Close()
			err = tcpHandler.Handle(c)
			if err != nil {
				log.Printf("handler exception: %s", err)
			}
		}(conn)
	}
}

func printGoroutines(log *log.Logger) {
	for {
		log.Printf("[Sys] Num of goroutines: %d", runtime.NumGoroutine())

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		log.Printf("[Sys] Heap sys Allocated mem: %.2f mb", float64(m.HeapSys)/1000/1000)
		log.Printf("[Sys] Heap Allocated mem: %.2f mb", float64(m.Alloc)/1000/1000)
		log.Printf("[Sys] GC circle num: %.2f ", float64(m.NumGC)/1000/1000)

		time.Sleep(5 * time.Second)
	}
}
