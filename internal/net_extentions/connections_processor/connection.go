package connections_processor

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	READ_TIMEOUT  = time.Second * 15
	WRITE_TIMEOUT = time.Second * 15
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

func (i *Connection) Read(cxt context.Context, l *log.Logger) {
	i.C.SetReadDeadline(time.Now().Add(READ_TIMEOUT))

	for {
		b, err := ReadMessage(i.C)

		if err != nil {
			if errors.Is(err, io.EOF) {
				//"write 1"
			} else if errors.Is(err, os.ErrDeadlineExceeded) {
				l.Printf("Connection [%s] with internal timeouted to read: %s", i.Name, err)
			} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				l.Printf("Connection [%s] with error timeouted: %s", i.Name, err)
			} else {
				l.Printf("Connection [%s] unknown error: %s", i.Name, err)
			}
			return
		} else if b != nil && len(b.GetMessageBytes()) != 0 {
			i.RecievedMem += uint64(len(b.GetMessageBytes()))

			select {
			case i.Recieved <- b:
			case <-time.After(WRITE_TIMEOUT):
				b.Release()
				return
			}
		}
	}
}

func (i *Connection) Write(cxt context.Context, l *log.Logger) {
	i.C.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT))
	timer := time.NewTimer(WRITE_TIMEOUT)

	for {
		select {
		case mes, ok := <-i.ToSend:
			if !ok {
				return
			}

			wrote, err := i.C.Write(mes.GetMessageBytes())
			if err == nil {
				i.SendedMem += uint64(wrote)
			} else if errors.Is(err, os.ErrDeadlineExceeded) {
				l.Printf("Connection [%s] with internal timeouted to write: %s", i.Name, err)
			}
			mes.Release()

			timer.Reset(WRITE_TIMEOUT)
		case <-timer.C:
			return
		}
	}
}
