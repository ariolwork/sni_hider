package net_extentions

import (
	"log"
	"net"
	"sync"
	"tcp_sni_splitter/internal/net_extentions/connections_processor"
)

func StartDoubleWayContentThrow(name string, s net.Conn, t net.Conn, l *log.Logger, buf connections_processor.BufProcessor) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	sConnection := connections_processor.WrapConnectionWithCustomWg("localhost", s, wg)
	tConnection := &connections_processor.Connection{Name: name, C: t, Recieved: sConnection.ToSend, ToSend: sConnection.Recieved, SendingWg: wg}
	defer buf.Statistics().AddStatistic(sConnection)
	defer buf.Statistics().AddStatistic(tConnection)
	buf.ProceedConnections(sConnection, tConnection)
	wg.Wait()
}
