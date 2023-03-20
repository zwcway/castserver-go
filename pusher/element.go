package pusher

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"go.uber.org/zap"
)

type Element struct {
	power bool

	line   *speaker.Line
	buffer *stream.Samples
	chBuf  [audio.Channel_MAX]*stream.Samples
}

func (e *Element) Name() string {
	return "Pusher"
}

func (e *Element) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (e *Element) On() {
	e.power = true
}

func (e *Element) Off() {
	e.power = false
}

func (e *Element) IsOn() bool {
	return e.power
}

func (e *Element) Close() error {
	e.buffer = nil
	return nil
}

func (e *Element) initBuf(samples int) {
	e.buffer = stream.NewSamples(samples, e.line.Output)

	for i := range e.chBuf {
		e.chBuf[i] = nil
	}
	for i, ch := range e.buffer.Format.Channels() {
		e.chBuf[i] = e.buffer.ChannelSamples(ch)
	}
}

func (e *Element) Stream(samples *stream.Samples) {
	if !e.power {
		return
	}
	if e.buffer.NbSamples < samples.NbSamples || !e.buffer.Format.Layout.Equal(e.line.Layout()) {
		e.initBuf(samples.NbSamples)
	}
	var (
		chList = e.line.Channels()
		i      int
		c      int
		ch     audio.Channel
		from   []audio.Channel
	)

	for i = 0; i < len(chList); i++ {
		ch = chList[i]
		chList[i] = 0

		from = e.line.ChannelRoute(ch)
		c = samples.ChannelsCountBySlice(from)
		if c == 0 {
			continue
		}
		if e.chBuf[i] == nil {
			continue
		}
		e.chBuf[i].BeZero()

		c = samples.MixChannel(e.chBuf[i], from)
		if c == 0 {
			continue
		}

		chList[i] = ch
	}

	// 由于存在声道路由功能，如果先转码后路由，样本数据可能已经不是float64格式，不方便混合
	e.line.Resample.Stream(e.buffer)

	for i, ch = range chList {
		if !ch.IsValid() {
			continue
		}
		if e.chBuf[i] == nil {
			continue
		}
		e.PushToLineChannel(ch, e.chBuf[i])
	}
}

func (e *Element) PushToLineChannel(ch audio.Channel, samples *stream.Samples) {
	for _, sp := range e.line.SpeakersByChannel(ch) {
		// TODO 为防止转码耗时过长，克隆新的缓存，并放置后台转码和推送
		e.PushSpeaker(sp, samples)
	}
}

func (e *Element) PushSpeaker(sp *speaker.Speaker, samples *stream.Samples) {
	queue := sp.Queue
	if queue == nil {
		// log.Error("speaker not connected", zap.String("speaker", sp.String()))
		return
	}
	if len(queue) == cap(queue) {
		log.Error("send queue full", zap.Uint32("speaker", uint32(sp.Id)), zap.Int("size", len(queue)))
		return
	}

	sp.PipeLine.Stream(samples)

	buf := ServerPush{
		Ver:     1,
		Seq:     1,
		Time:    1,
		Samples: samples.RawData[0],
	}
	p, err := buf.Pack()
	if err != nil {
		return
	}

	data := make([]byte, p.DataSize())
	copy(data, p.Bytes())

	sp.Statistic.Queue += uint32(len(data))

	queue <- speaker.QueueData{Speaker: sp, Data: data}
}

func (e *Element) Sample(*float64, int, int) {}

func NewElement(line *speaker.Line) stream.SwitchElement {
	return &Element{
		line: line,
	}
}
