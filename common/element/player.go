package element

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
)

type playerSamples struct {
	samples   *stream.Samples
	layout    audio.Layout
	pos       int
	resampled bool
}

func (p *playerSamples) IsPlaying() bool {
	return p.pos < p.samples.RequestNbSamples
}

func (p *playerSamples) CanRemove() bool {
	return p.samples == nil || !p.IsPlaying()
}

type Player struct {
	resample stream.ResampleElement

	streamers []*playerSamples

	format audio.Format
}

func (v *Player) Name() string {
	return "Test Player"
}

func (v *Player) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (p *Player) Stream(samples *stream.Samples) {
	if len(p.streamers) == 0 {
		return
	}

	var (
		mixed = 0
		i     = 0
		s     *playerSamples
	)

	if p.format != samples.Format {
		p.format = samples.Format
		go p.resampleAll()
	}

	for i = 0; i < len(p.streamers); i++ {
		s = p.streamers[i]

		if s.CanRemove() {
			utils.SliceQuickRemove(&p.streamers, i)
			i--
			continue
		}
		if !s.resampled {
			// 等待转码完成
			continue
		}

		if s.layout.IsValid() {
			// 混合至指定声道
			i = s.samples.MixChannels(samples, s.layout.Channels(), 0, s.pos)
		} else {
			// 按默认规则混合
			i = s.samples.MixChannelMap(samples, 0, s.pos)
		}
		s.pos += i
		if i == 0 {
			s.pos = s.samples.RequestNbSamples
		}
		if mixed < i {
			mixed = i
		}
	}

	samples.LastNbSamples = mixed
}

func (v *Player) Sample(sample *float64, ch int, n int) {}

func (p *Player) resampleAll() {
	if !p.format.IsValid() {
		return
	}
	for _, s := range p.streamers {
		if s.samples == nil || s.resampled || p.resample == nil {
			continue
		}

		f := p.format

		if s.layout.IsValid() {
			// 不改变播放源的声道布局
			f.Layout = s.layout
		}
		if s.samples.Format == f {
			s.resampled = true
			continue
		}

		f.InitFrom(s.samples.Format)

		p.resample.SetFormat(f)
		// 预先全部转码
		p.resample.Stream(s.samples)

		s.resampled = true
	}
}

func (p *Player) AddPCMWithChannel(ch audio.Channel, f audio.Format, pcm []byte) {
	if len(pcm) == 0 {
		return
	}
	f.Layout = audio.NewLayout(ch)
	p.streamers = append(p.streamers, &playerSamples{
		samples: stream.NewFromBytes(pcm, f),
		layout:  f.Layout,
	})

	p.resampleAll()
}

func (p *Player) AddPCM(f audio.Format, pcm []byte) {
	if len(pcm) == 0 {
		return
	}
	p.streamers = append(p.streamers, &playerSamples{
		samples: stream.NewFromBytes(pcm, f),
		layout:  audio.NewLayout(audio.Channel_NONE),
	})

	p.resampleAll()
}

func (p *Player) Close() (err error) {
	bus.UnregisterObj(p)

	if p.resample != nil {
		p.resample.Off()
		err = p.resample.Close()
	}
	return
}

func (o *Player) Dispatch(e string, a ...any) error {
	return bus.DispatchObj(o, e, a...)
}
func (o *Player) Register(e string, c bus.Handler) *bus.HandlerData {
	return bus.RegisterObj(o, e, c)
}

func (o *Player) AudioFormat() audio.Format {
	return o.format
}

func (o *Player) SetOutFormat(audio.Format) error {
	return nil
}

func (o *Player) CanRemove() bool {
	return len(o.streamers) == 0
}

func (o *Player) IsPlaying() bool {
	return len(o.streamers) != 0
}

func (o *Player) ChannelIndex() *audio.ChannelIndex {
	return o.format.ChannelIndex()
}

func newPlayer() stream.RawPlayerElement {
	var resample stream.ResampleElement

	p := &Player{}

	f := audio.DefaultFormat()
	stream.BusResample.GetInstance(p, &resample, &f)

	if resample != nil {
		resample.On()
	}

	p.resample = resample
	return p
}

func NewPlayer() stream.RawPlayerElement {
	p := newPlayer()

	return p
}
