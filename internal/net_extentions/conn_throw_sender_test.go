package net_extentions

import (
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

type mock_conn struct {
	rnd rand.Rand
}

func (c *mock_conn) Read(b []byte) (n int, err error) {
	n = int(c.rnd.Int31n(10000))
	for i := 0; i < n; i++ {
		b[i] = byte(i & 0xff)
	}
	time.Sleep(time.Duration(c.rnd.Int31n(50) * 1000))
	return n, err
}

func (c *mock_conn) Write(b []byte) (n int, err error)  { return 0, nil }
func (c *mock_conn) Close() error                       { return nil }
func (c *mock_conn) LocalAddr() net.Addr                { return nil }
func (c *mock_conn) RemoteAddr() net.Addr               { return nil }
func (c *mock_conn) SetDeadline(t time.Time) error      { return nil }
func (c *mock_conn) SetReadDeadline(t time.Time) error  { return nil }
func (c *mock_conn) SetWriteDeadline(t time.Time) error { return nil }

func BenchmarkMessageReadWithPool(b *testing.B) {
	wg := &sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		c := &mock_conn{*rand.New(rand.NewSource(int64(b.N)))}
		wg.Add(1)
		go func() {
			defer wg.Done()
			intwg := &sync.WaitGroup{}
			for k := 0; k < 200; k++ {
				intwg.Add(1)
				go func() {
					defer intwg.Done()
					mes, _ := ReadMessage(c)
					mes.Release()
				}()
				time.Sleep(time.Duration(c.rnd.Int31n(500) * 1000))
			}
			intwg.Wait()
		}()
	}
	wg.Wait()
}

func BenchmarkMessageReadWithoutPool(b *testing.B) {
	wg := &sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		c := &mock_conn{*rand.New(rand.NewSource(int64(b.N)))}
		wg.Add(1)
		go func() {
			defer wg.Done()
			intwg := &sync.WaitGroup{}
			for k := 0; k < 200; k++ {
				intwg.Add(1)
				go func() {
					defer intwg.Done()
					mes, _ := readMessageWithoutPool(c)
					mes.Release()
				}()
				time.Sleep(time.Duration(c.rnd.Int31n(20) * 1000))
			}
			intwg.Wait()
		}()
	}
	wg.Wait()
}
