package net_extentions

import (
	"io"
	"log"
	"net"
	"sync"
)

var memPool sync.Pool = sync.Pool{New: func() any { return make([]byte, BATCHSIZE) }}

const (
	BATCHSIZE = 16700 // tls message size 16384 + auth bytes 256
)

func dropContentThrow(s net.Conn, t net.Conn, l *log.Logger) error {
	memThrow := 0.0
	for {
		b, err := ReadMessage(s)
		if err != nil {
			if err == io.EOF {
				l.Printf("connection closed by user, traffic: %.2f kb", memThrow/1024)
				return nil
			} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				l.Printf("connection timeout, traffic: %.2f kb", memThrow/1024)
				return nil
			}
			l.Printf("connection error and closed %s, mem traffic: %.2f kb", err, memThrow/1024)
			return err
		}
		memThrow += float64(len(b.Message()))
		if len(b.Message()) != 0 {
			t.Write(b.Message())
		}
		b.Release()
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

func ReadMessage(s net.Conn) (Mess, error) {
	b := newMess()
	target := b.poolMem
	if target == nil {
		target = make([]byte, BATCHSIZE)
	}
	n, err := s.Read(target)
	if err == nil {
		b.message = target[:n]
		return b, nil
	}
	return nil, err
}
