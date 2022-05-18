package rx

import "sync"

type BehaiviorSubject struct {
	value []interface{}
	mu    sync.RWMutex
	ch    chan struct{}
}

func (bs *BehaiviorSubject) Publish(value interface{}) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.value = append(bs.value, value)
	if len(bs.value) == 0 {
		close(bs.ch)
	}
}

func (bs *BehaiviorSubject) Subscribe() []interface{} {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	<-bs.ch

	return bs.value
}
