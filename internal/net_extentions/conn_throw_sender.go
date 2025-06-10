package net_extentions

import (
	"io"
	"log"
	"net"
	"sync"
)

const (
	BATCHSIZE = 16700 // tls message size 16384 + auth bytes 256
)

func dropContentThrow(s net.Conn, t net.Conn, l *log.Logger) error {
	for {
		b, err := ReadMessage(s)
		if err != nil {
			if err == io.EOF {
				l.Printf("connection closed by user")
				return nil
			} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				l.Printf("connection timeout")
				return nil
			}
			l.Printf("connection error and closed %s", err)
			return err
		}
		if len(b) != 0 {
			t.Write(b)
		}
	}
}

func StartDoubleWayContentThrow(s net.Conn, t net.Conn, l *log.Logger) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		dropContentThrow(s, t, l)
	}()

	go func() {
		defer wg.Done()
		dropContentThrow(t, s, l)
	}()
	return wg
}

func ReadMessage(s net.Conn) ([]byte, error) {
	b := make([]byte, BATCHSIZE)
	n, err := s.Read(b)
	if err == nil {
		b = b[:n]
		return b, nil
	}
	return nil, err
}
