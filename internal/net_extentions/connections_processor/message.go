package connections_processor

import (
	"net"
	"sync"
)

const (
	BATCHSIZE = 16700 // tls message size 16384 + auth bytes 256
)

var messagePool *sync.Pool = &sync.Pool{New: func() any { return &message{b: make([]byte, BATCHSIZE)} }}

type Message interface {
	GetMessageBytes() []byte
	Release()
}

type message struct {
	b []byte
	m []byte
}

func (m *message) GetMessageBytes() []byte {
	return m.m
}

func (m *message) Release() {
	messagePool.Put(m)
}

func ReadMessage(s net.Conn) (Message, error) {
	r := messagePool.Get()
	mes := r.(*message)
	n, err := s.Read(mes.b)
	if err == nil {
		mes.m = mes.b[:n]
		return mes, nil
	}
	return nil, err
}
