package handlers

import (
	"errors"
	"fmt"
	"log"
	"net"
	"tcp_sni_splitter/internal/enumerable"
	"tcp_sni_splitter/internal/net_extentions"
	"tcp_sni_splitter/internal/net_extentions/handler_buf"
)

type tcpHandler struct {
	l                 *log.Logger
	connProcessorsBuf handler_buf.BufProcessor
}

func NewTcpHandler(l *log.Logger) Handler {
	return &tcpHandler{l: l, connProcessorsBuf: handler_buf.New()}
}

func (h *tcpHandler) Handle(c net.Conn) error {
	b, err := net_extentions.ReadMessage(c)
	if enumerable.IsStartFrom(b, []byte("CONNECT")) {
		if err = net_extentions.SetOK(c); err != nil {
			return nil
		}
	} else {
		// unknown first message for new connections
		return nil
	}
	targetPeer, err := net_extentions.ExtractTargetPeer(b)
	if err != nil {
		return errors.New(fmt.Sprintf("failing extract target peer address %s", err))
	}
	remoteConn, err := net.Dial("tcp", targetPeer.GetTargetUrl())
	if err != nil {
		return errors.New(fmt.Sprintf("failing to open target tcp %s", err))
	}
	defer remoteConn.Close()

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
	b, err := net_extentions.ReadMessage(s)
	if err != nil {
		return errors.New(fmt.Sprintf("failing rto read segment body %s", err))
	}
	segmented := net_extentions.SplitTLSBySegments(b)
	t.Write(segmented)
	return nil
}
