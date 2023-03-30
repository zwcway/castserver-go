package utils

import (
	"sync"
)

type WaitGroup struct {
	wg sync.WaitGroup
}

func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

func (w *WaitGroup) Go(cb func()) {
	w.wg.Add(1)
	go w.routine(cb)
}

func (w *WaitGroup) routine(cb func()) {
	cb()
	w.wg.Done()
}
