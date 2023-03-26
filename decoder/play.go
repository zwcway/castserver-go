package decoder

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/playlist"
	"github.com/zwcway/castserver-go/decoder/ffmpeg"
)

func Init() {
	bus.Register("get audioinfo", func(a ...any) error {
		ai := a[0].(*playlist.AudioInfo)

		return ffmpeg.AudioInfo(ai.Url, ai)
	})
}
