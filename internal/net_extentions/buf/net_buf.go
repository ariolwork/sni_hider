package buf

//todo описать собственный пул чтобы быстрее
import "sync"

const (
	BATCHSIZE     = 16700 // tls message size 16384 + auth bytes 256
	PARTITIONSIZE = BATCHSIZE * 64
	TAKENMASK     = 0xffffffffffffffff
)

type partition struct {
	b                 []byte
	takenMask         uint64
	pmu               *sync.Mutex
	id                int
	nextFreePartition *partition
}

type netBuf struct {
	nextFreePartition *partition
	partitions        []*partition
	m                 *sync.Mutex
}

func makePartition() *partition {
	return &partition{b: make([]byte, PARTITIONSIZE), takenMask: 0, pmu: &sync.Mutex{}}
}

func New() NetBuffer {
	return &netBuf{partitions: make([]*partition, 0, 100)}
}

// yes, unlimited grouth, but hardly imagine to keep more than 6400 connection in standart user experience
func (b *netBuf) Get() *NetSlice {
	if b.nextFreePartition == nil {
		b.m.Lock()
		if b.nextFreePartition == nil {
			newPartition := makePartition()
			newPartition.id = len(b.partitions)
			b.partitions = append(b.partitions, newPartition)
			b.nextFreePartition = newPartition
		}
		b.m.Unlock()
	}
	return b.getSclice()
}

func (b *netBuf) getSclice() *NetSlice {
	if b.nextFreePartition.takenMask != TAKENMASK {
		b.nextFreePartition.pmu.Lock()
		if b.nextFreePartition.takenMask != TAKENMASK {
			for i := 0; i < 64; i++ {
				takeMsk := (uint64(1) << i)
				if takeMsk&b.nextFreePartition.takenMask == 0 {
					b.nextFreePartition.takenMask = takeMsk | b.nextFreePartition.takenMask
					partition := b.nextFreePartition
					// no free space
					if partition.takenMask == TAKENMASK {
						b.m.Lock()
						b.nextFreePartition = partition.nextFreePartition
						b.m.Unlock()
					}
					partition.pmu.Unlock()
					return &NetSlice{B: partition.b[i*BATCHSIZE : (i+1)*BATCHSIZE], offset: i, partition: partition.id}
				}
			}
		}
	}
	return &NetSlice{make([]byte, PARTITIONSIZE), -1, -1}
}

func (b *netBuf) Release(s *NetSlice) {
	if s.partition == -1 {
		return
	}
	partition := b.partitions[s.partition]
	partition.pmu.Lock()
	partition.takenMask ^= uint64(1) << s.offset
	if partition.takenMask == TAKENMASK {
		b.m.Lock()
		if b.nextFreePartition == nil {
			b.nextFreePartition = partition
		} else {
			p := b.nextFreePartition
			for p.nextFreePartition != nil {
				p = p.nextFreePartition
			}
			p.nextFreePartition = partition
		}
		b.m.Unlock()
	}
	partition.pmu.Unlock()
}
