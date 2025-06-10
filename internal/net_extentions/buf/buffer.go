package buf

type NetSlice struct {
	B         []byte
	offset    int
	partition int
}

type NetBuffer interface {
	Get() *NetSlice
	Release(s *NetSlice)
}
