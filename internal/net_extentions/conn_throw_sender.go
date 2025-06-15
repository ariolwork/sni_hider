package net_extentions

import (
	"log"
	"net"
	"sync"
)

const (
	BATCHSIZE = 16700 // tls message size 16384 + auth bytes 256
)

func StartDoubleWayContentThrow(s net.Conn, t net.Conn, l *log.Logger, buf BufProcessor) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	sConnection := WrapConnectionWithCustomWg(s, wg)
	tConnection := &Connection{C: t, Recieved: sConnection.ToSend, ToSend: sConnection.Recieved, Wg: wg}
	buf.ProceedConnection(sConnection)
	buf.ProceedConnection(tConnection)
	return wg
}
