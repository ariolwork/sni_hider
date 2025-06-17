package net_extentions

import (
	"log"
	"net"
	"sync"
	"tcp_sni_splitter/internal/net_extentions/connections_processor"
)

func StartDoubleWayContentThrow(s net.Conn, t net.Conn, l *log.Logger, buf connections_processor.BufProcessor) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	sConnection := connections_processor.WrapConnectionWithCustomWg(s, wg)
	tConnection := &connections_processor.Connection{C: t, Recieved: sConnection.ToSend, ToSend: sConnection.Recieved, Wg: wg}
	buf.ProceedConnection(sConnection)
	buf.ProceedConnection(tConnection)
	return wg
}
