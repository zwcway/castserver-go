package element

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/utils"
)

type playerStreamer struct {
	pos int
	pcm []float64
}

type Player struct {
	pcm []playerStreamer
}

func (v *Player) Name() string {
	return "Test Player"
}

func (v *Player) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (v *Player) Stream(samples *stream.Samples) {
	if len(v.pcm) == 0 {
		return
	}
	format := samples.Format

	if format.SampleRate != audio.AudioRate_44100 || format.SampleBits != audio.AudioBits_32LEF {
		fmt.Println(format.String())
		// TODO 转码 pcm 后播放
		return
	}

	for s := 0; s < len(v.pcm); s++ {
		pcm := &v.pcm[s]

		for i := 0; i < samples.LastSize; i++ {
			if pcm.pos >= len(pcm.pcm) {
				// 播放完毕，移除
				if utils.SliceQuickRemove(&v.pcm, s) {
					s--
				}
				break
			}

			for ch := 0; ch < samples.Format.Layout.Count; ch++ {
				samples.Buffer[ch][i] += pcm.pcm[pcm.pos]
			}
			pcm.pos++
		}
	}

}

func (v *Player) Sample(sample *float64, ch int, n int) {}

func (p *Player) IsIdle() bool {
	return len(p.pcm) == 0
}

func (p *Player) Len() int {
	return len(p.pcm)
}

func (p *Player) Add(pcm []float64) {
	if pcm == nil {
		return
	}
	p.pcm = append(p.pcm, playerStreamer{pcm: pcm})
}

func NewPlayer() stream.RawPlayerElement {
	return &Player{pcm: make([]playerStreamer, 0)}
}
