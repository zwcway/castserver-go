package element

import (
	"github.com/zwcway/castserver-go/decoder"
	"golang.org/x/exp/slices"
)

// 音频混合器
type Mixer struct {
	streamers []decoder.Streamer
}

const MixerName = "Mixer"

func (m *Mixer) Name() string {
	return MixerName
}

func (m *Mixer) Type() decoder.ElementType {
	return decoder.ET_WholeSamples
}

func (m *Mixer) Len() int {
	return len(m.streamers)
}

func (m *Mixer) Add(s ...decoder.Streamer) {
	m.streamers = append(m.streamers, s...)
}

func (m *Mixer) Del(s decoder.Streamer) {
	idx := slices.Index(m.streamers, s)
	if idx <= 0 {
		return
	}
	m.streamers[idx] = m.streamers[len(m.streamers)-1]
	m.streamers = m.streamers[:len(m.streamers)-1]
}

func (m *Mixer) Has(s decoder.Streamer) bool {
	idx := slices.Index(m.streamers, s)
	return idx >= 0
}

func (m *Mixer) HasFileStreamer() decoder.FileStreamer {
	for _, s := range m.streamers {
		if s, ok := s.(decoder.FileStreamer); ok {
			return s
		}
	}
	return nil
}

func (m *Mixer) Clear() {
	m.streamers = m.streamers[:0]
}

func (m *Mixer) Stream(samples *decoder.Samples) {
	if len(m.streamers) == 0 {
		return
	}

	var tmp = decoder.NewSamples(512, samples.Format)

	j := 0
	for j < samples.Size {
		mixed := 0
		for si := 0; si < len(m.streamers); si++ {
			// 混合音频流
			m.streamers[si].Stream(tmp)
			for ch := 0; ch < tmp.Format.Layout.Count; ch++ {
				for i := 0; i < tmp.LastSize; i++ {
					samples.Buffer[ch][j+i] += tmp.Buffer[ch][i]
				}
			}
			if mixed < tmp.LastSize {
				mixed = tmp.LastSize
			}
			if tmp.LastErr != nil {
				// 移除出有问题的音频流
				sj := len(m.streamers) - 1
				m.streamers[si] = m.streamers[sj]
				m.streamers = m.streamers[:sj]
				si--
			}
		}
		j += mixed
	}
}

func (m *Mixer) Sample(*float64, int, int) {}

func NewMixer(streamers ...decoder.Streamer) *Mixer {
	m := &Mixer{}
	m.Add(streamers...)
	return m
}
