package handlers

import (
	"fmt"
	"log"
	"net"
	"tcp_sni_splitter/internal/enumerable"
	"tcp_sni_splitter/internal/net_extentions"
)

type tcpHandler struct {
	l *log.Logger
}

func NewTcpHandler(l *log.Logger) Handler {
	return &tcpHandler{l: l}
}

func (h *tcpHandler) Handle(c net.Conn) error {
	b, err := net_extentions.ReadMessage(c)
	if enumerable.IsStartFrom(b.Message(), []byte("CONNECT")) {
		if err = net_extentions.SetOK(c); err != nil {
			b.Release()
			return nil
		}
	} else {
		b.Release()
		// unknown first message for new connections
		return nil
	}

	targetPeer, err := net_extentions.ExtractTargetPeer(b.Message())
	b.Release()
	if err != nil {
		return fmt.Errorf("failing extract target peer address %s", err)
	}
	remoteConn, err := net.Dial("tcp", targetPeer.GetTargetUrl())
	if err != nil {
		return fmt.Errorf("failing to open target tcp %s", err)
	}
	defer remoteConn.Close()

	if net_extentions.IsHttps(targetPeer) {
		err := dropBySegments(c, remoteConn)
		if err != nil {
			return fmt.Errorf("failed to drop segmented tls content %s", err)
		}
	}

	wg := net_extentions.StartDoubleWayContentThrow(c, remoteConn, h.l)
	wg.Wait()
	return nil
}

func dropBySegments(s net.Conn, t net.Conn) error {
	b, err := net_extentions.ReadMessage(s)
	if err != nil {
		b.Release()
		return fmt.Errorf("failing rto read segment body %s", err)
	}
	segmented := net_extentions.SplitTLSBySegments(b.Message())
	t.Write(segmented)
	b.Release()
	return nil
}
