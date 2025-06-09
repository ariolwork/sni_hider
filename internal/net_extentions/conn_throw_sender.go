package net_extentions

import (
	"io"
	"log"
	"net"
	"sync"
)

func dropContentThrow(s net.Conn, t net.Conn, l *log.Logger) error {
	for {
		b := make([]byte, 17000)
		n, err := s.Read(b)

		if err != nil {
			if err == io.EOF {
				l.Printf("connection successfully closed")
				return nil
			} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				l.Printf("connection timeout")
				return nil
			}
			return err
		}
		b = b[:n]
		if len(b) != 0 {
			t.Write(b)
		}
	}
	return nil
}

func StartDoubleWayContentThrow(s net.Conn, t net.Conn, l *log.Logger) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() error {
		defer wg.Done()
		return dropContentThrow(s, t, l)
	}()

	go func() error {
		defer wg.Done()
		return dropContentThrow(s, t, l)
	}()
	return wg
}
