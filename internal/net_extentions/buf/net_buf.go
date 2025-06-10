package buf

//todo описать собственный пул чтобы быстрее
import "sync"

const (
	BATCHSIZE     = 16700 // tls message size 16384 + auth bytes 256
	PARTITIONSIZE = BATCHSIZE * 64
)

type partition struct {
	b         []byte
	takenMask uint64
	pmu       sync.Mutex
}

type netBuf struct {
	partitions []partition
	mu         sync.Pool
}

func New() NetBuffer {
	return &netBuf{}
}

func (b *netBuf) Get() *NetSlice {
	return nil
}

func (b *netBuf) Release(s *NetSlice) {

}
