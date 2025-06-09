package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"tcp_sni_splitter/internal/enumerable"
	"tcp_sni_splitter/internal/net_extentions"
)

func main() {
	subscribeListener()
}

func subscribeListener() error {
	l, err := net.Listen("tcp", ":3020")
	if err != nil {
		fmt.Printf("Cannot subsribe on port %s", err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Cannot accept connrection %s", err)
		}
		go func(c net.Conn) {
			err = handler(c)
			if err != nil {
				fmt.Println(err)
			}
			defer c.Close()
		}(conn)
	}
}

func handler(c net.Conn) error {
	b := make([]byte, 9000)
	n, err := c.Read(b)
	if err != nil || n == 0 {
		fmt.Printf("failing request processing %s", err)
	}
	if enumerable.IsStartFrom(b, []byte("CONNECT")) {
		_, err = c.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		if err != nil {
			return nil
		}
	} else {
		return nil
	}
	targetPeer, err := net_extentions.ExtractTargetPeer(b)
	if err != nil {
		fmt.Printf("failing extract target peer address %s", err)
	}
	remoteConn, err := net.Dial("tcp", targetPeer.GetTargetUrl())
	if err != nil {
		fmt.Printf("failing to open target tcp %s", err)
	}
	defer remoteConn.Close()

	if net_extentions.IsHttps(targetPeer) {
		dropBySegments(c, remoteConn)
	}
	l := log.New(os.Stdout, "dpi", 0)
	wg := net_extentions.StartDoubleWayContentThrow(c, remoteConn, l)
	wg.Wait()
	return nil
}

func dropBySegments(s net.Conn, t net.Conn) error {
	header := make([]byte, 5)
	n, err := s.Read(header)
	if n != 5 || err != nil {
		return fmt.Errorf("failed to read segment header")
	}

	newContent := make([]byte, 0, 9000)
	b := make([]byte, 9000)
	n, err = s.Read(b)
	if err != nil || n == 0 {
		fmt.Printf("failing rto read segment body %s", err)
	}
	b = b[:n]

	rnd := *rand.New(rand.NewSource(99))
	for len(b) > 0 {
		l := rnd.Intn(len(b))
		if l < 2 { // to escape too small pqackages
			l = len(b)
		}
		mc, _ := hex.DecodeString("1603")
		newContent = append(newContent, mc...)
		newContent = append(newContent, byte(rnd.Intn(10)))
		newContent = binary.BigEndian.AppendUint16(newContent, uint16(l))
		newContent = append(newContent, b[0:l]...)
		b = b[l:]
	}
	t.Write(newContent)
	return nil
}
