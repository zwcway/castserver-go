package utils

import "sync"

type WaitGroup struct {
	sync.WaitGroup
}

func (w *WaitGroup) Wrap(cb func()) {
	w.Add(1)
	go w.routine(cb)
}

func (w *WaitGroup) routine(cb func()) {
	cb()
	w.Done()
}
