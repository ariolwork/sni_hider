package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"tcp_sni_splitter/internal/enumerable"
	"tcp_sni_splitter/internal/net_extentions"
	"tcp_sni_splitter/internal/net_extentions/connections_processor"
	"tcp_sni_splitter/internal/net_extentions/dns"
)

const (
	TIMEBETWEENSTATISTICS = 30000000000
)

type tcpHandler struct {
	l   *log.Logger
	dns dns.Resolver
}

func NewTcpHandler(l *log.Logger, includeStatistica bool) Handler {
	handler := &tcpHandler{l: l, dns: dns.New(l)}

	return handler
}

func (h *tcpHandler) Handle(c net.Conn) error {
	// proceed new connection first message
	b, err := connections_processor.ReadMessage(c)
	if err != nil {
		return err
	}
	if enumerable.IsStartFrom(b.GetMessageBytes(), []byte("CONNECT")) {
		if err = net_extentions.SetOK(c); err != nil {
			return nil
		}
	} else {
		// unknown first message for new connections
		return nil
	}
	targetPeer, err := net_extentions.ExtractTargetPeer(context.Background(), b.GetMessageBytes(), h.dns)
	if err != nil {
		return fmt.Errorf("failing extract target peer address %s", err)
	}
	b.Release()

	// open target connection
	remoteConn, err := net.Dial("tcp", targetPeer.GetTargetUrl())
	if err != nil {
		return fmt.Errorf("failing to open target tcp %s", err)
	}
	defer remoteConn.Close()

	//define is https handshake message
	if net_extentions.IsHttps(targetPeer) {
		err := dropBySegments(c, remoteConn)
		if err != nil {
			return fmt.Errorf("failed to drop segmented tls content %s", err)
		}
	}

	net_extentions.StartDoubleWayContentThrow(targetPeer.GetTargetName(), c, remoteConn, h.l, connections_processor.NewDefault(h.l))
	return nil
}

func dropBySegments(s net.Conn, t net.Conn) error {
	b, err := connections_processor.ReadMessage(s)
	if err != nil {
		return errors.New(fmt.Sprintf("failing rto read segment body %s", err))
	}
	defer b.Release()
	segmented := net_extentions.SplitTLSBySegments(b.GetMessageBytes())
	t.Write(segmented)
	return nil
}
