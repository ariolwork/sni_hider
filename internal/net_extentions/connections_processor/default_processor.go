package connections_processor

import (
	"context"
	"log"
)

type defaultProcessors struct {
	l    *log.Logger
	stat Statistics //it is ok for local development, but Grafana and Prometheus shoul be much more better)
}

func (b *defaultProcessors) Cancel() {
	return
}

func (b *defaultProcessors) ProceedConnection(i *Connection) {
	go func() {
		i.Write(context.Background(), b.l)
		i.SendingWg.Done()
	}()

	go func() {
		i.Read(context.Background(), b.l)
		close(i.Recieved)
	}()
}

func (b *defaultProcessors) ProceedConnections(inf ...*Connection) {
	for _, i := range inf {
		go func() {
			i.Write(context.Background(), b.l)
			i.SendingWg.Done()
		}()

		go func() {
			i.Read(context.Background(), b.l)
			close(i.Recieved)
		}()
	}
}

func (b *defaultProcessors) Statistics() Statistics {
	return b.stat
}

func NewDefault(l *log.Logger) BufProcessor {
	return &defaultProcessors{l, NewStatisticsMonitor()}
}
