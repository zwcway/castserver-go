package pipeline

import (
	"sync"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
)

type PipeLineStreamer struct {
	stream stream.Element
	cost   time.Duration
}

func (s *PipeLineStreamer) Name() string {
	return s.stream.Name()
}

func (s *PipeLineStreamer) Cost() int {
	return int(s.cost.Milliseconds())
}

type PipeLine struct {
	buffer       *stream.Samples
	format       audio.Format
	wholeStreams []*PipeLineStreamer
	oneStreams   []*PipeLineStreamer

	cost    time.Duration
	maxCost time.Duration
	locker  sync.Mutex
}

func (p *PipeLine) Len() int {
	return len(p.wholeStreams) + len(p.oneStreams)
}

func (p *PipeLine) Buffer() *stream.Samples {
	return p.buffer
}

func (p *PipeLine) SetBuffer(sample *stream.Samples) {
	p.buffer = sample
}

func (p *PipeLine) Prepend(s stream.Element) {
	ps := []*PipeLineStreamer{{
		stream: s,
		cost:   0,
	}}

	// if s.Type() == stream.ET_OneSample {
	// 	p.oneStreams = append(ps, p.wholeStreams...)
	// } else if s.Type() == stream.ET_WholeSamples {
	p.wholeStreams = append(ps, p.wholeStreams...)
	// }
}

// TODO 防止循环引用
func (p *PipeLine) Append(s ...stream.Element) {
	for _, ss := range s {
		if ss == nil {
			continue
		}
		ps := &PipeLineStreamer{
			stream: ss,
			cost:   0,
		}
		// if ss.Type() == stream.ET_OneSample {
		// 	p.oneStreams = append(p.oneStreams, ps)
		// } else if ss.Type() == stream.ET_WholeSamples {
		p.wholeStreams = append(p.wholeStreams, ps)
		// }
	}
}

func (p *PipeLine) Clear() {
	p.wholeStreams = p.wholeStreams[:0]
	p.oneStreams = p.oneStreams[:0]
}

func (p *PipeLine) Format() audio.Format {
	return p.format
}

func (p *PipeLine) Close() error {
	for _, s := range p.wholeStreams {
		if sc, ok := s.stream.(stream.StreamCloser); ok {
			sc.Close()
		}
	}
	p.wholeStreams = nil
	return nil
}

func (p *PipeLine) Stream(sample *stream.Samples) {
	if sample == nil && p.buffer == nil {
		return
	}

	p.locker.Lock()
	defer p.locker.Unlock()

	buf := sample
	if buf == nil {
		buf = p.buffer
	}

	var t time.Time
	for _, s := range p.wholeStreams {
		t = time.Now()
		s.stream.Stream(buf)
		s.cost = time.Since(t)
	}

	// t = time.Now()
	// for ch := 0; ch < p.buffer.Format.Layout.Count; ch++ {
	// 	for i := 0; i < p.buffer.Size; i++ {
	// 		for _, s := range p.oneStreams {
	// 			s.stream.Sample(&p.buffer.Buffer[ch][i], ch, i)
	// 		}
	// 	}
	// }
	// for _, s := range p.oneStreams {
	// 	s.cost = time.Since(t)
	// }
	p.cost = time.Since(t)
	if p.cost > p.maxCost {
		p.maxCost = p.cost
	}
}

func (p *PipeLine) LastCost() time.Duration {
	return p.cost
}

func (p *PipeLine) LastMaxCost() time.Duration {
	return p.maxCost
}

func (p *PipeLine) Streamers() []*PipeLineStreamer {
	return append(p.oneStreams, p.wholeStreams...)
}

func (p *PipeLine) Lock() {
	p.locker.Lock()
}
func (p *PipeLine) Unlock() {
	p.locker.Unlock()
}

func NewPipeLine(format audio.Format, eles ...stream.Element) stream.PipeLiner {
	p := &PipeLine{
		wholeStreams: make([]*PipeLineStreamer, 0),
		oneStreams:   make([]*PipeLineStreamer, 0),
		format:       format,
	}

	p.Append(eles...)
	return p
}
