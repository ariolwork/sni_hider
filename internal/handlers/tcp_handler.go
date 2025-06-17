package handlers

import (
	"errors"
	"fmt"
	"log"
	"net"
	"tcp_sni_splitter/internal/enumerable"
	"tcp_sni_splitter/internal/net_extentions"
	"tcp_sni_splitter/internal/net_extentions/connections_processor"
)

type tcpHandler struct {
	l                 *log.Logger
	connProcessorsBuf connections_processor.BufProcessor
}

func NewTcpHandler(l *log.Logger) Handler {
	return &tcpHandler{l: l, connProcessorsBuf: connections_processor.New()}
}

func (h *tcpHandler) Handle(c net.Conn) error {
	// proceed new connection first message
	b, err := connections_processor.ReadMessage(c)
	if enumerable.IsStartFrom(b.GetMessageBytes(), []byte("CONNECT")) {
		if err = net_extentions.SetOK(c); err != nil {
			return nil
		}
	} else {
		// unknown first message for new connections
		return nil
	}
	targetPeer, err := net_extentions.ExtractTargetPeer(b.GetMessageBytes())
	if err != nil {
		return errors.New(fmt.Sprintf("failing extract target peer address %s", err))
	}
	b.Release()

	// open target connection
	remoteConn, err := net.Dial("tcp", targetPeer.GetTargetUrl())
	if err != nil {
		return errors.New(fmt.Sprintf("failing to open target tcp %s", err))
	}
	defer remoteConn.Close()

	//define is https handshake message
	if net_extentions.IsHttps(targetPeer) {
		err := dropBySegments(c, remoteConn)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to drop segmented tls content %s", err))
		}
	}

	wg := net_extentions.StartDoubleWayContentThrow(c, remoteConn, h.l, h.connProcessorsBuf)
	wg.Wait()
	return nil
}

func dropBySegments(s net.Conn, t net.Conn) error {
	b, err := connections_processor.ReadMessage(s)
	defer b.Release()
	if err != nil {
		return errors.New(fmt.Sprintf("failing rto read segment body %s", err))
	}
	segmented := net_extentions.SplitTLSBySegments(b.GetMessageBytes())
	t.Write(segmented)
	return nil
}
