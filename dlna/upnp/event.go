package upnp

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Subscribe struct {
	Sid     string
	URLs    []*url.URL
	Timeout time.Time
}

type Event struct {
	subscribers map[string]*Subscribe
}

func (e *Event) Subscribe(uuid string, callback []*url.URL, timeout time.Time) *Subscribe {

	if e.subscribers == nil {
		e.subscribers = make(map[string]*Subscribe)
	}

	sb := &Subscribe{
		Sid:     uuid,
		URLs:    callback,
		Timeout: timeout,
	}
	e.subscribers[uuid] = sb

	return sb
}

var (
	callbackRegexp = regexp.MustCompile(`<([^<>]+)>`)
)

func ParseCallback(callback string) (ret []*url.URL, err error) {
	list := callbackRegexp.FindAllStringSubmatch(callback, -1)
	for _, match := range list {
		var url *url.URL
		url, err = url.Parse(match[1])
		if err != nil {
			return
		}

		ret = append(ret, url)
	}
	return
}

func ParseTimeout(timeout string) (ret time.Time, err error) {
	times := strings.Split(timeout, "-")
	if len(times) == 2 {
		if times[0] == "Second" {
			var ti int64
			ti, err = strconv.ParseInt(times[1], 0, 32)
			if err != nil {
				return
			}
			ret = time.Now().Add(time.Duration(ti) * time.Second)
			return
		}
	}

	ret = time.Now()
	return
}
