package weak_hashring

import (
	"github.com/stathat/consistent"
	"sync"
)

type Hasher struct {
	Ring *consistent.Consistent
	mu   sync.Mutex
}

func NewHasher() *Hasher {
	return &Hasher{
		Ring: consistent.New(),
	}
}

func (h *Hasher) UpdateNodes(nodes []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Ring.Set(nodes)
}

func (h *Hasher) GetNodeForKey(key string) (string, error) {
	return h.Ring.Get(key)
}
