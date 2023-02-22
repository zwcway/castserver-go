package pipeline

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder"
)

type pipeLineStreamer struct {
	stream decoder.Element
	cost   time.Duration
}

type PipeLine struct {
	line      *speaker.Line
	buffer    *decoder.Samples
	format    *audio.Format
	streamers []pipeLineStreamer
	times     []time.Duration
}

func (p *PipeLine) Len() int {
	return len(p.streamers)
}
func (p *PipeLine) Buffer() *decoder.Samples {
	return p.buffer
}

func (p *PipeLine) Prepend(s decoder.Element) {
	p.streamers = append([]pipeLineStreamer{{
		stream: s,
		cost:   0,
	}}, p.streamers...)
	p.times = append(p.times, 0)
}

func (p *PipeLine) Add(s ...decoder.Element) {
	for _, ss := range s {
		p.streamers = append(p.streamers, pipeLineStreamer{
			stream: ss,
			cost:   0,
		})
	}
	p.times = append(p.times, 0)
}

func (p *PipeLine) Clear() {
	p.streamers = p.streamers[:0]
}

func (p *PipeLine) Format() *audio.Format {
	return p.format
}

func (p *PipeLine) Close() error {
	for _, s := range p.streamers {
		if sc, ok := s.stream.(decoder.StreamCloser); ok {
			sc.Close()
		}
	}
	return nil
}

func (p *PipeLine) Stream() (int, error) {
	var t time.Time
	for _, s := range p.streamers {
		t = time.Now()

		if s.stream.Type() == decoder.ET_WholeSamples {
			s.stream.Stream(p.buffer)
		} else if s.stream.Type() == decoder.ET_OneSample {
			for ch := 0; ch < p.buffer.Format.Layout.Count; ch++ {
				for i := 0; i < p.buffer.Size; i++ {
					s.stream.Sample(&p.buffer.Buffer[ch][i], ch, i)
				}
			}
		}

		s.cost = time.Since(t)
	}

	return p.buffer.LastSize, p.buffer.LastErr
}

func NewPipeLine(line *speaker.Line) *PipeLine {
	return &PipeLine{
		line: line,
	}
}
