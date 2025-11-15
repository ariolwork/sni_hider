package connections_processor

import (
	"context"
	"log"
	"sync"
)

const (
	GOROUTINESBUF = 1000
)

type processorsBuf struct {
	readCh  chan *Connection
	writeCh chan *Connection
	stat    Statistics //it is ok for local development, but Grafana and Prometheus shoul be much more better)

	updateChansMutex *sync.Mutex
	cancelChan       context.Context
	wg               *sync.WaitGroup
	cancel           func()
}

type BufProcessor interface {
	Cancel()
	ProceedConnection(inf *Connection)
	ProceedConnections(inf ...*Connection)
	Statistics() Statistics
}

func (b *processorsBuf) Cancel() {
	b.cancel()
	b.wg.Wait()
}

func (b *processorsBuf) ProceedConnection(inf *Connection) {
	b.updateChansMutex.Lock()
	defer b.updateChansMutex.Unlock()
	b.readCh <- inf
	b.writeCh <- inf
}

func (b *processorsBuf) ProceedConnections(inf ...*Connection) {
	b.updateChansMutex.Lock()
	defer b.updateChansMutex.Unlock()
	for _, i := range inf {
		b.readCh <- i
		b.writeCh <- i
	}
}

func (b *processorsBuf) Statistics() Statistics {
	return b.stat
}

func New(l *log.Logger) BufProcessor {
	context, cancel := context.WithCancel(context.Background())
	item := &processorsBuf{make(chan *Connection, GOROUTINESBUF*4), make(chan *Connection, GOROUTINESBUF*4), NewStatisticsMonitor(), &sync.Mutex{}, context, &sync.WaitGroup{}, cancel}
	item.wg.Add(GOROUTINESBUF * 2)
	for i := 0; i < GOROUTINESBUF; i++ {
		go func() {
			defer item.wg.Done()
			for {
				select {
				case <-item.cancelChan.Done():
					return
				case i := <-item.writeCh:
					i.Write(context, l)
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
					i.Read(context, l)
					close(i.Recieved)
				}
			}
		}()
	}
	return item
}
