package connections_processor

import (
	"sync"
	"sync/atomic"
)

type Conn_statistics struct {
	Name     string
	Recieved *uint64
	Sended   *uint64
}

type statisticMonitor struct {
	stats map[string]*Conn_statistics
	m     *sync.Mutex
}

type Statistics interface {
	AddStatistic(c *Connection)
	GetStatistic() []*Conn_statistics
}

func NewStatisticsMonitor() Statistics {
	return &statisticMonitor{make(map[string]*Conn_statistics, 100), &sync.Mutex{}}
}

func (s *statisticMonitor) AddStatistic(c *Connection) {
	var existsItem *Conn_statistics = nil
	var e bool = false
	if existsItem, e = s.stats[c.Name]; !e {
		s.m.Lock()
		if existsItem, e = s.stats[c.Name]; !e {
			existsItem = &Conn_statistics{c.Name, new(uint64), new(uint64)}
			s.stats[c.Name] = existsItem
		}
		s.m.Unlock()
	}
	if existsItem != nil {
		atomic.AddUint64(existsItem.Recieved, c.RecievedMem)
		atomic.AddUint64(existsItem.Sended, c.SendedMem)
	} else {
		panic("cannot add new statistic record to monitor")
	}
}

func (s *statisticMonitor) GetStatistic() []*Conn_statistics {
	s.m.Lock()
	defer s.m.Unlock()
	result := make([]*Conn_statistics, 0, len(s.stats))
	for _, v := range s.stats {
		a := new(uint64)
		*a = *a + *v.Recieved
		b := new(uint64)
		*b = *b + *v.Sended
		result = append(result, &Conn_statistics{v.Name, a, b})
	}
	return result
}
