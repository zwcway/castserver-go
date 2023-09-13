package playlist

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
)

type AudioInfo struct {
	Format   audio.Format
	Position time.Duration
	Duration time.Duration
	Url      string

	Title  string
	Artist string
}

type PlayList struct {
	list    []*AudioInfo
	current int
}

func (pl *PlayList) Add(url string) {
	ai := &AudioInfo{Url: url}

	bus.DispatchObj(pl, "get audioinfo", ai, len(pl.list))

	if ai.Duration > 0 {
		pl.list = append(pl.list, ai)
	}
}

func (pl *PlayList) PlayUrl(url string) {
	for i, ai := range pl.list {
		if ai.Url == url {
			pl.current = i

			bus.DispatchObj(pl, "playlist current changed", ai, i)
			return
		}
	}
}
