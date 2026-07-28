package element

import (
	"sync"

	lg "github.com/zwcway/castserver-go/common/log"

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
	tag string
	// 与顺序无关
	streamers []mixerStreamer

	buffer *stream.Samples

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

func (m *Mixer) Add(ss ...stream.SourceStreamer) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, s := range ss {
		if s == nil {
			continue
		}

		var resample stream.ResampleElement

		// stream.BusResample.GetInstance(m, &resample, nil)

		if resample != nil {
			resample.SetFormat(m.buffer.Format)
			resample.On()
		}

		m.streamers = append(m.streamers, mixerStreamer{
			s,
			resample,
		})

		ml.Debug("add source streamer", lg.String("tag", m.tag))

		// 订阅格式变更事件，同步调用
		stream.BusSourceFormatChanged.Register(s, m.onSourceFormatChanged)
	}

	return nil
}

func (m *Mixer) onSourceFormatChanged(ss stream.SourceStreamer, format *audio.Format, channelIndex audio.ChannelIndex) error {
	for _, s := range m.streamers {
		if s.streamer != ss {
			continue
		}

		if s.resample != nil {
			s.resample.OnFormatChanged(format)
		}
		stream.BusMixerInFormatChanged.Dispatch(m, ss, format, format.ChannelIndex())
	}

	return nil
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

func (m *Mixer) Format() audio.Format {
	return m.buffer.Format
}

func (m *Mixer) Stream(samples *stream.Samples) {
	if len(m.streamers) == 0 || m.buffer == nil {
		return
	}
	m.buffer.Resize(samples.RequestNbSamples, m.Format())

	mixed := 0
	for si := 0; si < len(m.streamers); si++ {
		ms := m.streamers[si]

		// m.buffer.ResetData()

		// 从音源获取数据
		ms.streamer.Stream(m.buffer)

		// 转码到内部格式
		if ms.resample != nil {
			ms.resample.Stream(m.buffer)
		}

		// 混合音源数据
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

func (m *Mixer) OnStarting() {
}

func (m *Mixer) OnEnding() {
}

func (m *Mixer) OnFormatChanged(newFormat *audio.Format) {
	f := *newFormat
	f.Bits = audio.Bits_DEFAULT

	for _, s := range m.streamers {
		if s.resample != nil {
			s.resample.SetFormat(f)
		}

		if f.LessThan(s.streamer.AudioFormat()) {
			f = s.streamer.AudioFormat()
		}
	}
	ml.Debug("format changed", lg.String("format", f.String()), lg.String("tag", m.tag))

	m.buffer = stream.NewSamplesDuration(config.AudioBuferMSDuration, f)

	stream.BusMixerOutFormatChanged.Dispatch(m, newFormat, f.ChannelIndex())
}

func (m *Mixer) Close() error {
	bus.UnregisterObj(m)

	for _, ss := range m.streamers {
		if ss.resample != nil {
			ss.resample.Close()
		}
		ss.streamer.Close()
	}

	m.Clear()
	return nil
}

func (m *Mixer) Dispatch(e string, a ...any) error {
	return bus.DispatchObj(m, e, a...)
}
func (m *Mixer) Register(e string, c bus.Handler) *bus.HandlerData {
	return bus.RegisterObj(m, e, c)
}

func NewMixer(tag string, streamers ...stream.SourceStreamer) stream.MixerElement {
	m := &Mixer{
		tag:       tag,
		streamers: make([]mixerStreamer, 0),
		buffer:    stream.NewSamplesDuration(config.AudioBuferMSDuration, audio.DefaultFormat()),
	}
	m.Add(streamers...)
	return m
}
