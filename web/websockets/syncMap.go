package websockets

import (
	"sync"

	"github.com/fasthttp/websocket"
)

type syncMap struct {
	sync.RWMutex
	m map[*websocket.Conn][]broadcastEvent
}

func (m *syncMap) del(key *websocket.Conn) {
	m.Lock()
	delete(m.m, key)
	m.Unlock()
}

func (m *syncMap) foreach(f func(key *websocket.Conn, val []broadcastEvent) bool) {
	m.RLock()
	defer m.RUnlock()

	for k, v := range m.m {
		if !f(k, v) {
			return
		}
	}
}

func (m *syncMap) get(key *websocket.Conn) (ret []broadcastEvent, ok bool) {
	m.RLock()
	ret, ok = m.m[key]
	m.RUnlock()
	return
}

func (m *syncMap) set(key *websocket.Conn, val []broadcastEvent) {
	m.Lock()
	if val == nil {
		m.m[key] = make([]broadcastEvent, 0)
	} else {
		m.m[key] = val
	}
	m.Unlock()
}

func createSyncMap() *syncMap {
	return &syncMap{m: make(map[*websocket.Conn][]broadcastEvent)}
}
