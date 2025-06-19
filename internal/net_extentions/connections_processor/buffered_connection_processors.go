package connections_processor

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
)

const (
	GOROUTINESBUF = 500
)

type Connection struct {
	Name        string
	ToSend      chan Message
	Recieved    chan Message
	SendedMem   uint64
	RecievedMem uint64

	C         net.Conn
	SendingWg *sync.WaitGroup
}

func WrapConnection(name string, c net.Conn) *Connection {
	return &Connection{Name: name, C: c, Recieved: make(chan Message, 10), ToSend: make(chan Message, 10), SendingWg: &sync.WaitGroup{}}
}

func WrapConnectionWithCustomWg(name string, c net.Conn, wg *sync.WaitGroup) *Connection {
	return &Connection{Name: name, C: c, Recieved: make(chan Message, 10), ToSend: make(chan Message, 10), SendingWg: wg}
}

type processorsBuf struct {
	readCh  chan *Connection
	writeCh chan *Connection
	stat    Statistics //it is ok for local development, but Grafana and Prometheus shoul be much more better)

	cancelChan context.Context
	wg         *sync.WaitGroup
	cancel     func()
}

type BufProcessor interface {
	Cancel()
	ProceedConnection(inf *Connection)
	Statistics() Statistics
}

func (b *processorsBuf) Cancel() {
	b.cancel()
	b.wg.Wait()
}

func (b *processorsBuf) ProceedConnection(inf *Connection) {
	b.readCh <- inf
	b.writeCh <- inf
}

func (b *processorsBuf) Statistics() Statistics {
	return b.stat
}

// обработку на чтение бы тоже в отдельную горутину

func New(l *log.Logger) BufProcessor {
	context, cancel := context.WithCancel(context.Background())
	item := &processorsBuf{make(chan *Connection, GOROUTINESBUF*4), make(chan *Connection, GOROUTINESBUF*4), NewStatisticsMonitor(), context, &sync.WaitGroup{}, cancel}
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
						wrote, err := i.C.Write(mes.GetMessageBytes())
						if err == nil {
							i.SendedMem += uint64(wrote)
						}
						mes.Release()
					}
					i.SendingWg.Done()
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
								l.Printf("Connection [%s] with error timeouted: %s", i.Name, err)
							} else {
								l.Printf("Connection [%s] unknown error: %s", i.Name, err)
							}
							close(i.Recieved)
							break
						} else if len(b.GetMessageBytes()) != 0 {
							i.RecievedMem += uint64(len(b.GetMessageBytes()))
							i.Recieved <- b
						}
					}
				}
			}
		}()
	}
	return item
}
