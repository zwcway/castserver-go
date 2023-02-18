package upnp

import (
	"net/url"
	"time"
)

type subscribe struct {
	urls []*url.URL
	timeout time.Time
}

type Event struct {
	subscribers map[string]*subscribe
}

func (e *Event) Subscribe(callback []*url.URL, timeout time.Time) {
	
}
