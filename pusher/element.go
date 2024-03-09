package pusher

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
)

type Element struct {
	power bool

	line   *speaker.Line
	buffer *stream.Samples
	chBuf  [audio.Channel_MAX]*stream.Samples

	resample stream.ResampleElement
}

func (e *Element) Name() string {
	return "Pusher"
}

func (e *Element) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (e *Element) On() {
	e.power = true
	e.resample.On()
}

func (e *Element) Off() {
	e.power = false
	e.resample.Off()
}

func (e *Element) IsOn() bool {
	return e.power
}

func (e *Element) Close() error {
	bus.UnregisterObj(e)

	if e.resample != nil {
		e.resample.Off()
		e.resample.Close()
	}
	e.buffer = nil
	return nil
}

func (e *Element) initBuf(samples int, format audio.Format) {
	if e.buffer == nil {
		e.buffer = stream.NewSamples(samples, format)
	} else if e.buffer.Format != format || samples != e.buffer.RequestNbSamples {
		e.buffer.Resize(samples, format)
	} else {
		e.buffer.ResetData()
		return
	}

	e.buffer.ResetData()
	e.buffer.Format = format
	e.resetData()
}

func (e *Element) resetData() {
	for i := range e.chBuf {
		e.chBuf[i] = nil
	}
	for i, ch := range e.buffer.Format.Channels() {
		samples := e.chBuf[i]
		if samples == nil {
			e.chBuf[i] = e.buffer.ChannelSamples(ch)
		} else {
			e.chBuf[i] = e.buffer.ChannelSamples(ch)
		}
	}
}

func (e *Element) Stream(samples *stream.Samples) {
	if !e.power || !samples.Format.IsValid() || !e.line.Layout().IsValid() {
		return
	}
	if samples.LastNbSamples == 0 {
		return
	}
	var (
		chList = e.line.Channels()
		i      int
		c      int
		ch     audio.Channel
		from   []audio.Channel
		buf    *stream.Samples
		format = samples.Format
		layout = samples.Format.Layout
	)
	format.Layout = e.line.Layout()

	e.initBuf(samples.RequestNbSamples, format)

	// 开始声道路由/混合
	for i = 0; i < len(chList); i++ {
		ch = chList[i]
		chList[i] = 0

		if e.chBuf[i] == nil {
			continue
		}

		from = e.line.ChannelRoute(ch)
		layout = samples.Format.Layout
		if !layout.IntersectSlice(from) {
			// 批量判断声道是否存在
			continue
		}

		buf = e.chBuf[i]
		buf.ResetData()

		// 混合多个声道至目标声道
		c = samples.MixChannels(buf, from, 0, 0)
		if c == 0 {
			continue
		}
		buf.LastNbSamples = c
		buf.Format.Rate = samples.Format.Rate

		e.buffer.LastNbSamples = c
		e.buffer.Format.Rate = samples.Format.Rate

		for _, sp := range e.line.SpeakersByChannel(ch) {
			sp.PipeLine.Stream(buf)
		}

		chList[i] = ch
	}

	// 由于存在声道路由功能，如果先转码后路由，样本数据可能已经不是float64格式，不方便混合
	e.resample.Stream(e.buffer)

	for i, ch = range chList {
		if !ch.IsValid() {
			continue
		}
		buf = e.chBuf[i]
		if buf == nil || buf.LastNbSamples == 0 {
			continue
		}
		buf.Format.Sample = e.buffer.Format.Sample

		for _, sp := range e.line.SpeakersByChannel(ch) {
			// TODO 为防止转码耗时过长，克隆新的缓存，并放置后台转码和推送
			e.PushSpeaker(sp, buf)
		}
	}
}
func (e *Element) OnStarting() {
}

func (e *Element) OnEnding() {
}

func (e *Element) OnFormatChanged(newFormat *audio.Format) {
}

func (e *Element) PushSpeaker(sp *speaker.Speaker, samples *stream.Samples) {
	queue := sp.Queue
	if queue == nil || sp.Conn == nil {
		// log.Error("speaker not connected", lg.String("speaker", sp.String()))
		return
	}
	if len(queue) == cap(queue) {
		log.Error("send queue full", lg.Uint("speaker", uint64(sp.ID)), lg.Int("size", int64(len(queue))))
		return
	}

	// 填充slient实现延迟
	delay := sp.EqualizerEle.Delay()

	buf := ServerPush{
		Ver:      1,
		Compress: 0,
		Rate:     samples.Format.Rate,
		Bits:     samples.Format.Bits,
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

	// TODO 使用缓存池
	data := make([]byte, delayBufSize+p.DataSize())
	copy(data[delayBufSize:], p.Bytes())

	sp.Statistic.Queue += uint32(len(data))

	queue <- speaker.QueueData{Speaker: sp, Data: data}
}

func (e *Element) Sample(*float64, int, int) {}

func (o *Element) Dispatch(e string, args ...any) error {
	return bus.DispatchObj(o, e, args...)
}
func (o *Element) Register(e string, c func(o any, a ...any) error) *bus.HandlerData {
	return bus.RegisterObj(o, e, c)
}

func bufSizeWithDelay(delay time.Duration, f audio.Format) int {
	return int((delay * time.Duration(f.Rate.ToInt()*f.Bits.Size()) * time.Second) / (time.Microsecond * 100))
}

func NewElement(line *speaker.Line) stream.SwitchElement {
	e := &Element{
		line: line,
	}
	var resample stream.ResampleElement
	stream.BusResample.GetInstance(e, &resample, &line.Output)
	if resample == nil {
		return nil
	}

	e.resample = resample

	line.Register("line output changed", func(o any, a ...any) error {
		l := o.(*speaker.Line)
		e.resample.SetFormat(l.Output)
		return nil
	})
	return e
}
