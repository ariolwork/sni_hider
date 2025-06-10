package net_extentions

type Mess interface {
	Release()
	Message() []byte
}

type mess struct {
	poolMem []byte
	message []byte
}

func newMess() *mess {
	buf := memPool.Get()
	if v, ok := buf.([]byte); ok {
		return &mess{poolMem: v}
	}
	return &mess{}
}

func (m *mess) Release() {
	if m.poolMem != nil {
		memPool.Put(m.poolMem)
	}
}

func (m *mess) Message() []byte {
	return m.message
}
