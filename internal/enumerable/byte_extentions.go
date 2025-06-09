package enumerable

func IsStartFrom(container []byte, search []byte) bool {
	if len(container) < len(search) {
		return false
	}
	for i, item := range search {
		if item != container[i] {
			return false
		}
	}
	return true
}
