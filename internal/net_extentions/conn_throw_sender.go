package net_extentions

import (
	"io"
	"log"
	"net"
	"sync"
	"tcp_sni_splitter/internal/net_extentions/handler_buf"
)

const (
	BATCHSIZE = 16700 // tls message size 16384 + auth bytes 256
)

func dropContentThrow(s net.Conn, t net.Conn, l *log.Logger, buf handler_buf.BufProcessor) error {
	//todo а давайте ка брать каналы из пула
	sender := &handler_buf.Processing{C: t, ToSend: make(chan []byte, 10), Wg: &sync.WaitGroup{}}
	buf.Send(sender)
	sender.Wg.Add(1)
	defer sender.Wg.Wait()
	defer close(sender.ToSend)
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
			sender.ToSend <- b
		}
	}
}

func StartDoubleWayContentThrow(s net.Conn, t net.Conn, l *log.Logger, buf handler_buf.BufProcessor) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		dropContentThrow(s, t, l, buf)
	}()

	go func() {
		defer wg.Done()
		dropContentThrow(t, s, l, buf)
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
