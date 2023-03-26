package element

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
)

type playerStreamer struct {
	pos     int
	layout  audio.ChannelLayout
	samples *stream.Samples
}

type Player struct {
	resample stream.ResampleElement
	pcm      []playerStreamer
}

func (v *Player) Name() string {
	return "Test Player"
}

func (v *Player) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (v *Player) Stream(samples *stream.Samples) {
	if len(v.pcm) == 0 || v.resample == nil {
		return
	}
	format := samples.Format
	mixed := 0
	i := 0

	for s := 0; s < len(v.pcm); s++ {
		pcm := &v.pcm[s]

		if !pcm.samples.Format.SampleEqual(&format) {
			// 不改变播放源的声道布局
			if pcm.layout.IsValid() {
				format.Layout = pcm.layout
			} else {
				format.Layout = pcm.samples.Format.Layout
			}
			v.resample.SetFormat(format)
			// 转码程序会操作所有样本数据
			// 转码过后 pcm.samples.Format 会被赋值为 format
			// 因此只会被执行一次
			v.resample.Stream(pcm.samples)
		}
		if pcm.layout.IsValid() {
			i = pcm.samples.MixChannels(samples, pcm.layout.Channels(), 0, pcm.pos)
		} else {
			i = pcm.samples.MixChannelMap(samples, 0, pcm.pos)
		}
		pcm.pos += i
		if mixed < i {
			mixed = i
		}

		if i == 0 || pcm.pos >= pcm.samples.NbSamples {
			// 混合失败 或者 播放完毕，移除
			if utils.SliceQuickRemove(&v.pcm, s) {
				s--
			}
		}
	}
	if mixed > samples.LastNbSamples && mixed <= samples.NbSamples {
		samples.LastNbSamples = mixed
	}
}

func (v *Player) Sample(sample *float64, ch int, n int) {}

func (p *Player) IsIdle() bool {
	return len(p.pcm) == 0
}

func (p *Player) Len() (c int) {
	return len(p.pcm)
}

func (p *Player) AddToChannel(ch audio.Channel, f audio.Format, pcm []byte) {
	if len(pcm) == 0 {
		return
	}
	samples := stream.NewSamplesCopy(pcm, f)
	//
	p.pcm = append(p.pcm, playerStreamer{
		layout:  audio.NewChannelLayout(ch),
		samples: samples,
	})
}

func (p *Player) Add(f audio.Format, pcm []byte) {
	p.AddToChannel(audio.Channel_NONE, f, pcm)
}

func (p *Player) Close() (err error) {
	p.resample.Off()
	err = p.resample.Close()
	p.pcm = p.pcm[:0]
	return
}

func NewPlayer() stream.RawPlayerElement {
	resample := NewResample(audio.DefaultFormat())
	resample.On()

	return &Player{
		resample: resample,
		pcm:      make([]playerStreamer, 0),
	}
}
