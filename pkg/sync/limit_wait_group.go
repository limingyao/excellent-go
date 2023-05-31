package sync

import (
	"time"

	"go.uber.org/atomic"
)

type LimitWaitGroup struct {
	wg  WaitGroup
	ch  chan struct{}
	err atomic.Error
}

func NewLimitWaitGroup(concurrency int) *LimitWaitGroup {
	return &LimitWaitGroup{ch: make(chan struct{}, concurrency)}
}

func (wg *LimitWaitGroup) Add(delta int) {
	wg.wg.Add(1)
	for i := 0; i < delta; i++ {
		wg.ch <- struct{}{}
	}
}

func (wg *LimitWaitGroup) AddTimeout(timeout time.Duration) bool {
	wg.wg.Add(1)
	select {
	case wg.ch <- struct{}{}:
		return false
	case <-time.After(timeout):
		wg.wg.Done()
		return true
	}
}

// AddCheck 检查执行过程中是否出错，出错直接返回
func (wg *LimitWaitGroup) AddCheck(delta int) error {
	if err := wg.err.Load(); err != nil {
		return err
	}
	wg.Add(delta)
	return nil
}

func (wg *LimitWaitGroup) Done() {
	<-wg.ch
	wg.wg.Done()
}

func (wg *LimitWaitGroup) Wait() {
	wg.wg.Wait()
}

func (wg *LimitWaitGroup) WaitTimeout(timeout time.Duration) bool {
	return wg.wg.WaitTimeout(timeout)
}

// SetError 设置出错信息
func (wg *LimitWaitGroup) SetError(err error) {
	wg.err.Store(err)
}
