package limiting

import (
	"sync"
	"sync/atomic"
)

// Limiting limit goroutine num.
type Limiting struct {
	c   chan struct{}
	wg  *sync.WaitGroup
	err atomic.Value // 用于并发过程中出错快速退出
}

// NewLimiting ...
func NewLimiting(size int) *Limiting {
	if size <= 0 {
		panic("goroutine limiting size <= 0")
	}
	return &Limiting{
		c:  make(chan struct{}, size),
		wg: new(sync.WaitGroup),
	}
}

// Add adds delta to the WaitGroup counter.
func (g *Limiting) Add(delta int) {
	g.wg.Add(delta)
	for i := 0; i < delta; i = i + 1 {
		g.c <- struct{}{}
	}
}

// Done decrements the WaitGroup counter by one.
func (g *Limiting) Done() {
	<-g.c
	g.wg.Done()
}

// Wait blocks until the WaitGroup counter is zero.
func (g *Limiting) Wait() {
	g.wg.Wait()
}

// AddCheck adds delta to the WaitGroup counter.
// 用于快速退出
func (g *Limiting) AddCheck(delta int) error {
	if err := g.err.Load().(error); err != nil {
		return err
	}
	g.Add(delta)
	return nil
}

// Error ...
// 用于快速退出
func (g *Limiting) Error(err error) {
	if err != nil {
		g.err.Store(err)
	}
}
