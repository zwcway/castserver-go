package utils

import (
	"sync"
	"sync/atomic"
)

type WaitGroup struct {
	wg    sync.WaitGroup
	c     int32
	exitC chan struct{}
	lock  sync.Mutex
}

func (w *WaitGroup) Wait() {
	if w.c > 0 {
		w.wg.Wait()
	}
}

func (w *WaitGroup) Go(cb func(<-chan struct{})) {
	if w.exitC == nil {
		w.exitC = make(chan struct{})
	}
	atomic.AddInt32(&w.c, 1)
	w.wg.Add(1)
	go w.routine(cb)
}

func (w *WaitGroup) ExitAndWait() {
	if w.c == 0 || w.exitC == nil {
		return
	}
	w.lock.Lock()
	defer w.lock.Unlock()

	close(w.exitC)
	w.Wait()
	w.exitC = nil
}

func (w *WaitGroup) routine(cb func(<-chan struct{})) {
	cb(w.exitC)
	atomic.AddInt32(&w.c, -1)
	w.wg.Done()
}
