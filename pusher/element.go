package pusher

import (
	"time"

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

func (e *Element) initBuf(samples int, format audio.Format) {
	format.Layout = *e.line.Layout()
	if e.buffer == nil {
		e.buffer = stream.NewSamples(samples, format)
	} else {
		e.buffer.ResizeSamplesOrNot(samples, format)
	}

	for i := range e.chBuf {
		e.chBuf[i] = nil
	}
	for i, ch := range e.buffer.Format.Channels() {
		e.chBuf[i] = e.buffer.ChannelSamples(ch)
	}
}

func (e *Element) Stream(samples *stream.Samples) {
	if !e.power || !samples.Format.IsValid() || !e.line.Layout().IsValid() {
		return
	}
	if samples.LastNbSamples == 0 {
		return
	}
	if e.buffer == nil || e.buffer.NbSamples < samples.NbSamples || !e.buffer.Format.Layout.Equal(e.line.Layout()) {
		e.initBuf(samples.NbSamples, samples.Format)
	}
	var (
		chList = e.line.Channels()
		i      int
		c      int
		ch     audio.Channel
		from   []audio.Channel
		buf    *stream.Samples
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
		buf = e.chBuf[i]
		buf.ResetData()

		c = samples.MixChannels(buf, from, 0, 0)
		if c == 0 {
			continue
		}
		buf.LastNbSamples = c

		for _, sp := range e.line.SpeakersByChannel(ch) {
			sp.PipeLine.Stream(buf)
		}

		chList[i] = ch
	}

	// 由于存在声道路由功能，如果先转码后路由，样本数据可能已经不是float64格式，不方便混合
	e.line.ResampleEle.Stream(e.buffer)

	for i, ch = range chList {
		if !ch.IsValid() {
			continue
		}
		buf = e.chBuf[i]
		if buf == nil || buf.LastNbSamples == 0 {
			continue
		}
		for _, sp := range e.line.SpeakersByChannel(ch) {
			// TODO 为防止转码耗时过长，克隆新的缓存，并放置后台转码和推送
			e.PushSpeaker(sp, buf)
		}
	}
}

func (e *Element) PushSpeaker(sp *speaker.Speaker, samples *stream.Samples) {
	queue := sp.Queue
	if queue == nil {
		// log.Error("speaker not connected", zap.String("speaker", sp.String()))
		return
	}
	if len(queue) == cap(queue) {
		log.Error("send queue full", zap.Uint32("speaker", uint32(sp.ID)), zap.Int("size", len(queue)))
		return
	}

	delay := sp.EqualizerEle.Delay()

	buf := ServerPush{
		Ver:      1,
		Compress: 0,
		Rate:     samples.Format.SampleRate,
		Bits:     samples.Format.SampleBits,
		Time:     uint16(delay) + 1,
		Samples:  samples.ChannelBytes(0),
	}
	p, err := buf.Pack()
	if err != nil {
		return
	}

	if delayChanged(sp, delay) {
		// 更改了延迟时间，抛弃旧的队列，生成新的队列
		refreshPushQueue(sp, delay)
	}

	// 按照采样率填充指定大小的静音样本实现延迟指定时间
	delayBufSize := bufSizeWithDelay(delay, sp.Format())

	data := make([]byte, delayBufSize+p.DataSize())
	copy(data[delayBufSize:], p.Bytes())

	sp.Statistic.Queue += uint32(len(data))

	queue <- speaker.QueueData{Speaker: sp, Data: data}
}

func (e *Element) Sample(*float64, int, int) {}

func bufSizeWithDelay(delay time.Duration, f audio.Format) int {
	return int((delay * time.Duration(f.SampleRate.ToInt()*f.SampleBits.Size()) * time.Second) / (time.Microsecond * 100))
}

func NewElement(line *speaker.Line) stream.SwitchElement {
	return &Element{
		line: line,
	}
}
