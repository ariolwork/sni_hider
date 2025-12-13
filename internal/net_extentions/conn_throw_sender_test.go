package net_extentions

import (
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"tcp_sni_splitter/internal/net_extentions/connections_processor"
	"testing"
	"time"
)

type mock_conn struct {
	rnd            *rand.Rand
	sendingDelay   int
	recievingDelay int
	sendingData    int
}

func NewMockConnection(r *rand.Rand) *mock_conn {
	return &mock_conn{r, r.Intn(50) + 10, r.Intn(50) + 10, 500000}
}

func (c *mock_conn) Read(b []byte) (n int, err error) {
	if c.sendingData == 0 {
		return 0, io.EOF
	}
	n = int(c.rnd.Int31n(int32(c.sendingData)))
	if n > 16000 {
		n = 16000
	}
	if c.sendingData < 150 {
		n = c.sendingData
	}
	for i := 0; i < n; i++ {
		b[i] = byte(i & 0xff)
	}
	time.Sleep(time.Duration(c.recievingDelay * 1000))
	c.sendingData -= n
	return n, err
}

func (c *mock_conn) Write(b []byte) (n int, err error) {
	time.Sleep(time.Duration(c.sendingDelay * 1000))
	return len(b), nil
}

func (c *mock_conn) Close() error                       { return nil }
func (c *mock_conn) LocalAddr() net.Addr                { return nil }
func (c *mock_conn) RemoteAddr() net.Addr               { return nil }
func (c *mock_conn) SetDeadline(t time.Time) error      { return nil }
func (c *mock_conn) SetReadDeadline(t time.Time) error  { return nil }
func (c *mock_conn) SetWriteDeadline(t time.Time) error { return nil }

type mock_writer struct{}

func (m *mock_writer) Write(p []byte) (n int, err error) {
	return 0, nil
}

func BenchmarkDoubleWayContentThrowTest(b *testing.B) {
	l := log.New(&mock_writer{}, "", 0)
	wg := &sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		rnd := rand.New(rand.NewSource(int64(i)))
		c1 := NewMockConnection(rnd)
		c2 := NewMockConnection(rnd)
		wg.Add(1)
		go func() {
			StartDoubleWayContentThrow("test", c1, c2, l, connections_processor.NewDefault(l))
			wg.Done()
		}()
	}
	wg.Wait()
}
