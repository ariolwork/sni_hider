package net_extentions

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
)

var tlsHeader []byte = func() []byte {
	r, err := hex.DecodeString("160308")
	if err != nil {
		panic("bad tls const header")
	}
	return r
}()

func SplitTLSBySegments(b []byte) []byte {
	if len(b) < 6 {
		return b
	}
	//skip header bytes
	b = b[5:]
	if len(b) < 4 {
		//strange content, bad for seed
		return b
	}
	result := make([]byte, 0, BATCHSIZE)
	rnd := *rand.New(rand.NewSource(int64(binary.BigEndian.Uint32(b[:4]))))
	parts := 0
	for len(b) > 0 {
		l := rnd.Intn(len(b))
		if l < 50 && parts > 1 { // to escape too small parts
			l = len(b)
		}
		if parts > 5 { // to escape large parts amount
			l = len(b)
		}
		result = append(result, tlsHeader...)
		result = binary.BigEndian.AppendUint16(result, uint16(l))
		result = append(result, b[0:l]...)
		b = b[l:]
		parts++
	}
	return result
}
