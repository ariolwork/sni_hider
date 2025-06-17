package connections_processor

import (
	"context"
	"io"
	"net"
	"sync"
)

const (
	GOROUTINESBUF = 500
)

type Connection struct {
	ToSend   chan Message
	Recieved chan Message
	C        net.Conn
	Wg       *sync.WaitGroup
}

func WrapConnection(c net.Conn) *Connection {
	return &Connection{C: c, Recieved: make(chan Message, 10), ToSend: make(chan Message, 10), Wg: &sync.WaitGroup{}}
}

func WrapConnectionWithCustomWg(c net.Conn, wg *sync.WaitGroup) *Connection {
	return &Connection{C: c, Recieved: make(chan Message, 10), ToSend: make(chan Message, 10), Wg: wg}
}

type processorsBuf struct {
	readCh     chan *Connection
	writeCh    chan *Connection
	cancelChan context.Context
	wg         *sync.WaitGroup
	cancel     func()
}

type BufProcessor interface {
	Cancel()
	ProceedConnection(inf *Connection)
}

func (b *processorsBuf) Cancel() {
	b.cancel()
	b.wg.Wait()
}

func (b *processorsBuf) ProceedConnection(inf *Connection) {
	b.readCh <- inf
	b.writeCh <- inf
}

// обработку на чтение бы тоже в отдельную горутину

func New() BufProcessor {
	context, cancel := context.WithCancel(context.Background())
	item := &processorsBuf{make(chan *Connection, GOROUTINESBUF*4), make(chan *Connection, GOROUTINESBUF*4), context, &sync.WaitGroup{}, cancel}
	item.wg.Add(GOROUTINESBUF * 2)
	for i := 0; i < GOROUTINESBUF; i++ {
		go func() {
			defer item.wg.Done()
			for {
				select {
				case <-item.cancelChan.Done():
					return
				case i := <-item.writeCh:
					for mes := range i.ToSend {
						i.C.Write(mes.GetMessageBytes())
						mes.Release()
					}
					i.Wg.Done()
				}
			}
		}()

		go func() {
			defer item.wg.Done()
			for {
				select {
				case <-item.cancelChan.Done():
					return
				case i := <-item.readCh:
					for {
						b, err := ReadMessage(i.C)
						if err != nil {
							if err == io.EOF {
								//"write 1"
							} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
								//write 2
							}
							close(i.Recieved)
							break
						} else if len(b.GetMessageBytes()) != 0 {
							i.Recieved <- b
						}
					}
				}
			}
		}()
	}
	return item
}
