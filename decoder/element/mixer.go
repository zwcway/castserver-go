package element

import (
	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/utils"
	"golang.org/x/exp/slices"
)

// 音频混合器
type Mixer struct {
	streamers []stream.Streamer

	fileStreamer stream.FileStreamer

	bufSize int
	buffer  *stream.Samples
	bufPos  int
}

func (m *Mixer) Name() string {
	return "Mixer"
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

func (m *Mixer) PreAdd(ss ...stream.Streamer) {
	utils.SlicePrepend(&m.streamers, ss...)
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

	if m.buffer == nil || !m.buffer.Format.Equal(&samples.Format) {
		m.buffer = stream.NewSamples(m.bufSize, samples.Format)
	}

	nbSamples := 0

	for nbSamples < samples.NbSamples {
		if m.bufPos > 0 && m.bufPos < m.buffer.LastNbSamples {
			// 残留数据
			i := 0
			for ch := 0; ch < m.buffer.Format.Layout.Count && ch < samples.Format.Layout.Count; ch++ {
				for i = 0; m.bufPos+i < m.buffer.LastNbSamples && nbSamples+i < samples.NbSamples; i++ {
					samples.Data[ch][i] += m.buffer.Data[ch][m.bufPos+i]
				}
			}
			nbSamples += i
			m.bufPos += i
			continue
		}
		m.bufPos = 0
		mixed := 0
		for si := 0; si < len(m.streamers); si++ {
			stream := m.streamers[si]
			// TODO 处理在line删除后的情况

			m.buffer.BeZero()
			// 混合音频流
			stream.Stream(m.buffer)
			i := 0
			for ch := 0; ch < m.buffer.Format.Layout.Count && ch < samples.Format.Layout.Count; ch++ {
				for i = 0; i < m.buffer.LastNbSamples && i+nbSamples < samples.NbSamples; i++ {
					samples.Data[ch][nbSamples+i] += m.buffer.Data[ch][i]
				}
			}
			if mixed < i {
				mixed = i
			}
		}
		if mixed == 0 {
			break
		}
		nbSamples += mixed
		m.bufPos += mixed
	}

	if m.buffer.LastErr != nil {
		if samples.LastErr != nil {
			samples.LastErr = errors.Wrap(m.buffer.LastErr, samples.LastErr.Error())
		} else {
			samples.LastErr = m.buffer.LastErr
		}
	}

	if m.buffer.HasData {
		samples.HasData = true
	}

	samples.SetFormat(m.buffer.Format)
	if nbSamples > samples.LastNbSamples {
		samples.LastNbSamples = nbSamples
	}
}

func (m *Mixer) Sample(*float64, int, int) {}

func (e *Mixer) Close() error {
	for _, s := range e.streamers {
		if sc, ok := s.(stream.StreamCloser); ok {
			sc.Close()
		}
	}

	e.Clear()
	return nil
}

func NewMixer(streamers ...stream.Streamer) stream.MixerElement {
	m := NewEmptyMixer()
	m.Add(streamers...)
	return m
}

func NewEmptyMixer() stream.MixerElement {
	m := &Mixer{
		streamers: make([]stream.Streamer, 0),
		bufSize:   config.AudioBuferSize,
	}
	return m
}
