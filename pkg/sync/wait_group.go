package pool

import (
	"sync"
	"time"
)

// WaitGroup 增强版 sync.WaitGroup 支持 WaitTimeout
type WaitGroup struct {
	sync.WaitGroup
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{}
}

// WaitTimeout 支持超时 wait
func (wg *WaitGroup) WaitTimeout(timeout time.Duration) bool {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		wg.Wait()
	}()

	select {
	case <-ch: // completed normally
		return false
	case <-time.After(timeout): // timed out
		return true
	}
}
