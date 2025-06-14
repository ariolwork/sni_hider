package handler_buf

import (
	"context"
	"net"
	"sync"
)

const (
	GOROUTINESBUF = 200
)

type Processing struct {
	ToSend chan []byte
	C      net.Conn
	Wg     *sync.WaitGroup
}

type buf struct {
	ch         chan *Processing
	cancelChan context.Context
	wg         *sync.WaitGroup
	cancel     func()
}

type BufProcessor interface {
	Cancel()
	Send(inf *Processing)
}

func (b *buf) Cancel() {
	b.cancel()
	b.wg.Wait()
}

func (b *buf) Send(inf *Processing) {
	b.ch <- inf
}

// обработку на чтение бы тоже в отдельную горутину

func New() BufProcessor {
	context, cancel := context.WithCancel(context.Background())
	item := &buf{make(chan *Processing, GOROUTINESBUF*2), context, &sync.WaitGroup{}, cancel}
	item.wg.Add(GOROUTINESBUF * 2)
	for i := 0; i < GOROUTINESBUF; i++ {
		go func() {
			defer item.wg.Done()
			for {
				select {
				case <-item.cancelChan.Done():
					return
				case i := <-item.ch:
					for bytes := range i.ToSend {
						i.C.Write(bytes)
					}
					i.Wg.Done()
				}
			}
		}()
	}
	return item
}
