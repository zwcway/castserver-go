package element

import (
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/utils"
	"golang.org/x/exp/slices"
)

// 音频混合器
type Mixer struct {
	streamers []stream.Streamer

	fileStreamer stream.FileStreamer
}

const MixerName = "Mixer"

func (m *Mixer) Name() string {
	return MixerName
}

func (m *Mixer) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (m *Mixer) Len() int {
	return len(m.streamers)
}

func (m *Mixer) Add(ss ...stream.Streamer) {
	for _, s := range ss {
		if s == nil {
			continue
		}
		m.streamers = append(m.streamers, s)
	}
}

func (m *Mixer) AddFileStreamer(s stream.FileStreamer) {
	if m.fileStreamer != nil {
		return
	}
	m.fileStreamer = s
	m.Add(s)
}

func (m *Mixer) Del(s stream.Streamer) {
	utils.SliceQuickRemoveItem(&m.streamers, s)
}

func (m *Mixer) Has(s stream.Streamer) bool {
	idx := slices.Index(m.streamers, s)
	return idx >= 0
}

func (m *Mixer) FileStreamer() stream.FileStreamer {
	return m.fileStreamer
}

func (m *Mixer) Clear() {
	m.streamers = m.streamers[:0]
}

func (m *Mixer) Stream(samples *stream.Samples) {
	if len(m.streamers) == 0 {
		return
	}

	var tmp = stream.NewSamples(samples.Size, samples.Format)

	mixed := 0
	for si := 0; si < len(m.streamers); si++ {
		stream := m.streamers[si]
		// TODO 处理在line删除后的情况

		// 混合音频流
		stream.Stream(tmp)
		for ch := 0; ch < tmp.Format.Layout.Count && ch < samples.Format.Layout.Count; ch++ {
			for i := 0; i < tmp.LastSize && i < samples.Size; i++ {
				samples.Buffer[ch][i] += tmp.Buffer[ch][i]
			}
		}
		if mixed < tmp.LastSize {
			mixed = tmp.LastSize
		}
	}
	// // 复制所有的buffer
	// for ch := 0; ch < samples.Format.Layout.Count; ch++ {
	// 	for i := 0; i < samples.LastSize; i++ {
	// 		samples.Buffer[ch][i] = tmp.Buffer[ch][i]
	// 	}
	// }
	samples.Format = tmp.Format
	if mixed > samples.LastSize {
		samples.LastSize = mixed
	}
}

func (m *Mixer) Sample(*float64, int, int) {}

func NewMixer(streamers ...stream.Streamer) stream.MixerElement {
	m := &Mixer{streamers: make([]stream.Streamer, 0)}
	m.Add(streamers...)
	return m
}
