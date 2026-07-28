package pipeline

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/stream"
)

type Element struct {
	ele  stream.Element
	cost time.Duration
}

func (s *Element) Name() string {
	return s.ele.Name()
}

func (s *Element) Cost() time.Duration {
	return s.cost
}

func (s *Element) Element() stream.Element {
	return s.ele
}

type PipeLine struct {
	tag      string
	buffer   *stream.Samples
	format   audio.Format
	elements []*Element

	cost    time.Duration
	maxCost time.Duration
}

func (p *PipeLine) Len() int {
	return len(p.elements)
}

func (p *PipeLine) Buffer() *stream.Samples {
	return p.buffer
}

func (p *PipeLine) SetBuffer(sample *stream.Samples) {
	p.buffer = sample
}

func (p *PipeLine) Prepend(s stream.Element) {
	ps := []*Element{{
		ele:  s,
		cost: 0,
	}}

	// if s.Type() == stream.ET_OneSample {
	// 	p.oneStreams = append(ps, p.wholeStreams...)
	// } else if s.Type() == stream.ET_WholeSamples {
	p.elements = append(ps, p.elements...)
	// }
}

// TODO 防止循环引用
func (p *PipeLine) Append(s ...stream.Element) {
	for _, ss := range s {
		if ss == nil {
			continue
		}

		ps := &Element{
			ele:  ss,
			cost: 0,
		}
		// if ss.Type() == stream.ET_OneSample {
		// 	p.oneStreams = append(p.oneStreams, ps)
		// } else if ss.Type() == stream.ET_WholeSamples {
		p.elements = append(p.elements, ps)
		// }

		log.Debug("element added", lg.String("name", ss.Name()), lg.String("tag", p.tag))
		// 同步格式，各自初始化自己的缓存
		ss.OnFormatChanged(&p.format)
	}
}

func (p *PipeLine) Clear() {
	p.elements = p.elements[:0]
}

func (p *PipeLine) AudioFormat() audio.Format {
	return p.format
}

func (p *PipeLine) ChannelIndex() audio.ChannelIndex {
	if p.buffer == nil {
		return p.format.ChannelIndex()
	}
	return p.buffer.ChannelIndex
}

func (p *PipeLine) SetOutFormat(f audio.Format) error {
	if p.format == f {
		return nil
	}

	p.format = f

	log.Debug("set out format", lg.String("format", f.String()), lg.String("tag", p.tag))

	for _, ele := range p.elements {
		ele.ele.OnFormatChanged(&f)
	}
	return nil
}

func (p *PipeLine) IsPlaying() bool {
	return true
}

func (p *PipeLine) CanRemove() bool {
	return false
}

func (p *PipeLine) Close() error {
	for _, s := range p.elements {
		if sc, ok := s.ele.(stream.StreamCloser); ok {
			sc.Close()
		}
		bus.UnregisterObj(s)
	}
	p.Clear()
	return nil
}

func (p *PipeLine) Stream(sample *stream.Samples) {
	if sample == nil && p.buffer == nil || !p.format.IsValid() {
		return
	}

	buf := sample
	if buf == nil {
		buf = p.buffer
	}

	for _, s := range p.elements {
		s.ele.OnStarting()
	}

	var (
		t  time.Time
		mt time.Time = time.Now()
	)
	for _, s := range p.elements {
		t = time.Now()
		s.ele.Stream(buf)
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
	p.cost = time.Since(mt)
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

func (p *PipeLine) Streamers() []*Element {
	return append([]*Element{}, p.elements...)
}

func NewPipeLine(tag string, format audio.Format, eles ...stream.Element) stream.PipeLiner {
	p := &PipeLine{
		tag:      tag,
		elements: make([]*Element, 0),
		format:   format,
	}

	p.Append(eles...)
	return p
}
