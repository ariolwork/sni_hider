package handlers

import (
	"errors"
	"fmt"
	"log"
	"net"
	"tcp_sni_splitter/internal/enumerable"
	"tcp_sni_splitter/internal/net_extentions"
	"tcp_sni_splitter/internal/net_extentions/connections_processor"
	"time"
)

const (
	TIMEBETWEENSTATISTICS = 30000000000
)

type tcpHandler struct {
	l                 *log.Logger
	connProcessorsBuf connections_processor.BufProcessor
}

func NewTcpHandler(l *log.Logger, includeStatistica bool) Handler {
	handler := &tcpHandler{l: l, connProcessorsBuf: connections_processor.New(l)}
	//log statistic
	if includeStatistica {
		go func() {
			for {
				m := handler.connProcessorsBuf.Statistics().GetStatistic()
				l.Printf("-------------------%s-------------------", time.Now().Local().UTC())
				for _, item := range m {
					l.Printf("%s   |   recieved: %.2f kb,  sended: %.2f kb", item.Name, float64(*item.Recieved)*1.0/1000, float64(*item.Sended)*1.0/1000)
				}
				l.Println("-----------------------------------------------------")
				time.Sleep(TIMEBETWEENSTATISTICS)
			}
		}()
	}
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

	net_extentions.StartDoubleWayContentThrow(targetPeer.GetTargetName(), c, remoteConn, h.l, h.connProcessorsBuf)
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
