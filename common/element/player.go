package element

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/stream"
)

type Player struct {
	resample stream.ResampleElement

	pos     int
	layout  audio.Layout
	samples *stream.Samples

	resampled bool
}

func (v *Player) Name() string {
	return "Test Player"
}

func (v *Player) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (p *Player) Stream(samples *stream.Samples) {
	if p.CanRemove() {
		return
	}
	if !p.resampled {
		// 等待转码完成
		return
	}

	mixed := 0

	if p.layout.IsValid() {
		// 混合至指定声道
		mixed = p.samples.MixChannels(samples, p.layout.Channels(), 0, p.pos)
	} else {
		// 按默认规则混合
		mixed = p.samples.MixChannelMap(samples, 0, p.pos)
	}
	p.pos += mixed

	if mixed == 0 || !p.IsPlaying() {
		// 混合失败 或者 播放完毕
	}
	samples.LastNbSamples = mixed
	samples.Format = p.resample.Format()
}

func (v *Player) Sample(sample *float64, ch int, n int) {}

func (p *Player) IsPlaying() bool {
	return p.pos < p.samples.RequestNbSamples
}

func (p *Player) CanRemove() bool {
	return p.samples == nil || !p.IsPlaying()
}

func (p *Player) AudioFormat() audio.Format {
	return p.resample.Format()
}

func (p *Player) SetOutFormat(f audio.Format) error {
	if p.samples == nil {
		return nil
	}

	if p.layout.IsValid() {
		// 不改变播放源的声道布局
		f.Layout = p.layout
	}
	if p.samples.Format == f {
		p.resampled = true
		return nil
	}
	if p.resample == nil {
		p.samples = nil
		return nil
	}

	f.InitFrom(p.samples.Format)

	p.resample.SetFormat(f)
	// 预先全部转码
	p.resample.Stream(p.samples)

	p.resampled = true

	return nil
}

func (p *Player) SetPCMWithChannel(ch audio.Channel, f audio.Format, pcm []byte) {
	if len(pcm) == 0 {
		p.samples = nil
		return
	}
	f.Layout = audio.NewLayout(ch)
	p.samples = stream.NewFromBytes(pcm, f)
	p.layout = f.Layout
}

func (p *Player) SetPCM(f audio.Format, pcm []byte) {
	if len(pcm) == 0 {
		p.samples = nil
		return
	}
	p.samples = stream.NewFromBytes(pcm, f)
	p.layout = audio.NewLayout(audio.Channel_NONE)
}

func (p *Player) Close() (err error) {
	if p.resample != nil {
		p.resample.Off()
		err = p.resample.Close()
	}
	return
}

func newPlayer(f audio.Format, pcm []byte) stream.RawPlayerElement {
	var resample stream.ResampleElement

	p := &Player{}

	bus.Dispatch("get resample element", p, &resample, audio.DefaultFormat())

	if resample != nil {
		resample.On()
	}

	p.resample = resample
	return p
}

func NewPlayerChannel(ch audio.Channel, f audio.Format, pcm []byte) stream.RawPlayerElement {
	p := newPlayer(f, pcm)

	p.SetPCMWithChannel(ch, f, pcm)
	return p
}

func NewPlayer(f audio.Format, pcm []byte) stream.RawPlayerElement {
	p := newPlayer(f, pcm)

	p.SetPCM(f, pcm)
	return p
}
