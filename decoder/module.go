package decoder

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/playlist"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/decoder/ffmpeg"
)

var (
	Module = decoderModule{}
)

type decoderModule struct{}

func (decoderModule) Init(ctx utils.Context) error {
	bus.Register("get audioinfo", func(a ...any) error {
		ai := a[0].(*playlist.AudioInfo)

		return ffmpeg.AudioInfo(ai.Url, ai)
	})

	bus.Register("get resample element", func(a ...any) error {
		// from := a[0]
		resample := a[1].(*stream.ResampleElement)
		format := a[2].(audio.Format)

		*resample = NewResample(format)
		return nil
	})

	return nil
}

func (decoderModule) Start() error {
	return nil
}

func (decoderModule) DeInit() {

}
