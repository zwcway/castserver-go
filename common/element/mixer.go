package element

import (
	"sync"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
)

type mixerStreamer struct {
	streamer stream.SourceStreamer
	resample stream.ResampleElement
}

const (
	noResample uint8 = iota
	hasResample
	onResample
)

// 音频混合器
type Mixer struct {
	// 与顺序无关
	streamers []mixerStreamer

	buffer   *stream.Samples
	resample uint8

	locker sync.Mutex
}

func (m *Mixer) Name() string {
	return "Mixer"
}

func (m *Mixer) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (m *Mixer) Len() int {
	m.locker.Lock()
	defer m.locker.Unlock()

	return len(m.streamers)
}

func (m *Mixer) Add(ss ...stream.SourceStreamer) {
	m.locker.Lock()
	defer m.locker.Unlock()

	var resample stream.ResampleElement

	for _, s := range ss {
		if s == nil {
			continue
		}

		stream.BusResample.GetInstance(m, &resample, nil)

		if resample == nil {
			m.resample = noResample
		}
		m.streamers = append(m.streamers, mixerStreamer{
			s,
			resample,
		})

		// 订阅格式变更事件，同步调用
		stream.BusSourceFormatChanged.Register(s, m.onSourceFormatChanged)
	}

	m.decideFormat()
}

func (m *Mixer) onSourceFormatChanged(ss stream.SourceStreamer, format *audio.Format, channelIndex audio.ChannelIndex) error {
	m.decideFormat()
	return nil
}

// 根据每个待混合格式确定合适的统一格式
func (m *Mixer) decideFormat() {
	if len(m.streamers) == 0 {
		return
	}

	format := audio.InternalFormat()

	for i := 0; i < len(m.streamers); i++ {
		s := m.streamers[i].streamer
		f := s.AudioFormat()
		if format.Count < f.Count {
			format.Layout = f.Layout
		}
		if format.Rate.LessThan(f.Rate) {
			format.Rate = f.Rate
		}
	}

	m.SetFormat(format)
}

func (m *Mixer) Del(s stream.SourceStreamer) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for i := 0; i < len(m.streamers); i++ {
		if m.streamers[i].streamer == s {
			utils.SliceQuickRemove(&m.streamers, i)
			s.Close()
			return
		}
	}
}

func (m *Mixer) Has(s stream.SourceStreamer) bool {
	m.locker.Lock()
	defer m.locker.Unlock()

	for i := 0; i < len(m.streamers); i++ {
		if m.streamers[i].streamer == s {
			return true
		}
	}
	return false
}

func (m *Mixer) Clear() {
	m.locker.Lock()
	defer m.locker.Unlock()

	m.streamers = m.streamers[:0]
}

func (m *Mixer) Buffer() *stream.Samples {
	return m.buffer
}

func (m *Mixer) SetResample(on bool) {
	m.locker.Lock()
	defer m.locker.Unlock()

	if m.resample == noResample {
		return
	}

	if on {
		for _, ss := range m.streamers {
			ss.resample.On()
		}
		m.resample = onResample
	} else {
		for _, ss := range m.streamers {
			ss.resample.Off()
		}
		m.resample = hasResample
	}
}

func (m *Mixer) SetFormat(format audio.Format) {
	if format == m.buffer.Format {
		return
	}
	format.Bits = audio.Bits_DEFAULT

	m.buffer.ResizeDuration(config.AudioBuferMSDuration, format)
	stream.BusMixerFormatChanged.Dispatch(m, &format, m.buffer.ChannelIndex)

	if m.resample != onResample {
		// 通知所有输入源变更输出格式
		for _, ms := range m.streamers {
			f := ms.streamer.AudioFormat()
			f.Sample = format.Sample
			ms.streamer.SetOutFormat(f)
		}
	}

}

func (m *Mixer) Format() audio.Format {
	return m.buffer.Format
}

func (m *Mixer) Stream(samples *stream.Samples) {
	if len(m.streamers) == 0 {
		return
	}

	if m.buffer.LessThan(samples) {
		m.buffer.Resize(samples.RequestNbSamples, samples.Format)
	}

	mixed := 0
	for si := 0; si < len(m.streamers); si++ {
		ms := m.streamers[si]

		m.buffer.ResetData()

		ms.streamer.Stream(m.buffer)

		if m.resample == onResample && ms.resample != nil {
			ms.resample.Stream(m.buffer)
		}

		i := m.buffer.MixChannelMap(samples, 0, 0)

		if mixed < i {
			mixed = i
		}

		if ms.streamer.CanRemove() {
			m.Del(ms.streamer)
			si--
		}
	}

	samples.WrapError(m.buffer.LastErr)
	samples.LastNbSamples = mixed
	samples.Format = m.buffer.Format
	// if mixed > samples.LastNbSamples && mixed <= samples.RequestNbSamples {
	// 	samples.LastNbSamples = mixed
	// }
}

func (m *Mixer) Sample(*float64, int, int) {}

func (m *Mixer) Close() error {
	bus.UnregisterObj(m)

	for _, ss := range m.streamers {
		ss.streamer.Close()
	}

	m.Clear()
	return nil
}

func (o *Mixer) Dispatch(e string, a ...any) error {
	return bus.DispatchObj(o, e, a...)
}
func (o *Mixer) Register(e string, c bus.Handler) *bus.HandlerData {
	return bus.RegisterObj(o, e, c)
}

func NewMixer(streamers ...stream.SourceStreamer) stream.MixerElement {
	m := &Mixer{
		streamers: make([]mixerStreamer, 0),
		buffer:    stream.NewSamplesDuration(config.AudioBuferMSDuration, audio.DefaultFormat()),
	}
	m.Add(streamers...)
	return m
}
