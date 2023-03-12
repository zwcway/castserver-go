package pusher

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
)

type Element struct {
	power bool

	line   *speaker.Line
	buffer *stream.Samples

	layout audio.ChannelLayout
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

func (e *Element) Stream(samples *stream.Samples) {
	if !e.power {
		return
	}
	if e.buffer.NbSamples < samples.NbSamples || !e.layout.Equal(e.line.Layout()) {
		e.layout = e.line.Output.Layout
		e.buffer = stream.NewSamples(samples.NbSamples, e.line.Output)
	}

	chList := e.line.Channels()
	for i, ch := range chList {
		from := e.line.ChannelRoute(ch)
		c := samples.ChannelsCountBySlice(from)
		if c == 0 {
			continue
		}
		e.buffer.BeZero()

		e.buffer.Format.Layout = audio.NewChannelLayout(ch)
		size := samples.MixChannel(e.buffer.Data[i], from)
		if size == 0 {
			continue
		}
	}

	e.line.Resample.Stream(e.buffer)

	for i, ch := range chList {
		PushToLineChannel(e.line, ch, e.buffer.ChannelBytes(i))
	}
}

func (e *Element) Sample(*float64, int, int) {}

func NewElement(line *speaker.Line) stream.SwitchElement {
	return &Element{
		line: line,
	}
}
