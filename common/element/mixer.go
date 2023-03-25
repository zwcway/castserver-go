package element

import (
	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/utils"
	"golang.org/x/exp/slices"
)

// 音频混合器
type Mixer struct {
	streamers []stream.Streamer

	buffer *stream.Samples
	bufPos int
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

func (m *Mixer) Del(s stream.Streamer) {
	utils.SliceQuickRemoveItem(&m.streamers, s)
}

func (m *Mixer) Has(s stream.Streamer) bool {
	idx := slices.Index(m.streamers, s)
	return idx >= 0
}

func (m *Mixer) Clear() {
	m.streamers = m.streamers[:0]
}

func (m *Mixer) Buffer() *stream.Samples {
	return m.buffer
}

func (m *Mixer) Stream(samples *stream.Samples) {
	if len(m.streamers) == 0 {
		return
	}

	// 保证 buffer 大于等于请求的大小
	// 并且设置为请求的格式
	m.buffer.ResizeSamplesOrNot(samples.NbSamples, samples.Format)

	nbSamples := 0

	for nbSamples < samples.NbSamples {
		if m.bufPos > 0 && m.bufPos < m.buffer.LastNbSamples {
			// 残留数据
			i := m.buffer.MixChannelMap(samples, nbSamples, m.bufPos)
			nbSamples += i
			m.bufPos += i
			continue
		}
		m.bufPos = 0
		mixed := 0
		for si := 0; si < len(m.streamers); si++ {
			stream := m.streamers[si]

			m.buffer.ResetData()

			// 注意：内部有可能改变 buffer
			stream.Stream(m.buffer)

			samples.ResizeSamplesOrNot(m.buffer.LastNbSamples, m.buffer.Format)

			i := m.buffer.MixChannelMap(samples, nbSamples, 0)

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

	if nbSamples > samples.LastNbSamples && nbSamples <= samples.NbSamples {
		samples.ResizeSamplesOrNot(nbSamples, m.buffer.Format)
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
		buffer:    stream.NewSamples(config.AudioBuferSize, audio.DefaultFormat()),
	}
	return m
}
