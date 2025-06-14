package rec_processors

import (
	"context"
	"net"
	"sync"
)

const (
	GOROUTINESBUF = 150
)

type ToSend struct {
	B  []byte
	C  net.Conn
	Wg *sync.WaitGroup
}

type buf struct {
	ch         chan *ToSend
	cancelChan context.Context
	wg         *sync.WaitGroup
	cancel     func()
}

type Buf interface {
	Cancel()
	Send(inf *ToSend)
}

func (b *buf) Cancel() {
	b.cancel()
	b.wg.Wait()
}

func (b *buf) Send(inf *ToSend) {
	b.ch <- inf
}

func New() Buf {
	context, cancel := context.WithCancel(context.Background())
	item := &buf{make(chan *ToSend, GOROUTINESBUF*2), context, &sync.WaitGroup{}, cancel}
	item.wg.Add(GOROUTINESBUF)
	for i := 0; i < GOROUTINESBUF; i++ {
		go func() {
			defer item.wg.Done()
			for {
				select {
				case <-item.cancelChan.Done():
					return
				case i := <-item.ch:
					i.C.Write(i.B)
					i.Wg.Done()
				}
			}
		}()
	}

	return item
}
