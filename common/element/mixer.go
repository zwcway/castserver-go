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

// 音频混合器
type Mixer struct {
	// 与顺序无关
	streamers []mixerStreamer

	buffer   *stream.Samples
	resample bool

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

		bus.Dispatch("get resample element", m, &resample, audio.InternalFormat())

		if resample == nil {
			continue
		}
		m.streamers = append(m.streamers, mixerStreamer{
			s,
			resample,
		})

		// 订阅格式变更事件，同步调用
		bus.RegisterObj(s, "source format changed", func(a ...any) error {
			m.decideFormat()
			return nil
		})
	}

	m.decideFormat()
}

func (m *Mixer) Del(s stream.SourceStreamer) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for i := 0; i < len(m.streamers); i++ {
		if m.streamers[i].streamer == s {
			utils.SliceQuickRemove(&m.streamers, i)
			s.Close()
			bus.UnregisterObj(s)
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

	m.resample = on
	if m.resample {
		for i := 0; i < len(m.streamers); i++ {
			m.streamers[i].resample.On()
		}
	} else {
		for i := 0; i < len(m.streamers); i++ {
			m.streamers[i].resample.Off()
		}
	}
}

func (m *Mixer) SetFormat(format audio.Format) {
	if format == m.buffer.Format {
		return
	}
	m.buffer.ResizeDuration(config.AudioBuferMSDuration, format)
	bus.Dispatch("mixer format changed", m, &format)
}

func (m *Mixer) Format() audio.Format {
	return m.buffer.Format
}

// 根据每个待混合格式确定合适的统一格式
func (m *Mixer) decideFormat() {
	format := m.buffer.Format
	format.Bits = audio.Bits_DEFAULT

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

	if !m.resample {
		// 通知所有输入源变更输出格式
		for i := 0; i < len(m.streamers); i++ {
			m.streamers[i].streamer.SetOutFormat(m.buffer.Format)
		}
	}
}

func (m *Mixer) Stream(samples *stream.Samples) {
	if len(m.streamers) == 0 {
		return
	}

	// 确保buffer大小满足请求的大小
	m.buffer.Resize(samples.RequestNbSamples, samples.Format)
	m.buffer.RequestNbSamples = samples.RequestNbSamples

	mixed := 0
	for si := 0; si < len(m.streamers); si++ {
		ms := m.streamers[si]

		m.buffer.ResetData()

		ms.streamer.Stream(m.buffer)

		ms.resample.Stream(m.buffer)

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
	// if mixed > samples.LastNbSamples && mixed <= samples.RequestNbSamples {
	// 	samples.LastNbSamples = mixed
	// }
}

func (m *Mixer) Sample(*float64, int, int) {}

func (m *Mixer) Close() error {
	for _, ss := range m.streamers {
		ss.streamer.Close()
		bus.UnregisterObj(ss.streamer)
	}

	m.Clear()
	return nil
}

func (m *Mixer) Dispatch(e string, a ...any) error {
	a = append([]any{m}, a...)
	return bus.Dispatch(e, a...)
}
func (m *Mixer) Register(e string, c func(a ...any) error) *bus.HandlerData {
	return bus.RegisterObj(m, e, c)
}

func NewMixer(streamers ...stream.SourceStreamer) stream.MixerElement {
	m := &Mixer{
		streamers: make([]mixerStreamer, 0),
		buffer:    stream.NewSamplesDuration(config.AudioBuferMSDuration, audio.DefaultFormat()),
	}
	m.Add(streamers...)
	return m
}
