package pipeline

import (
	"sync"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder"
	"github.com/zwcway/castserver-go/decoder/element"
)

type PipeLineStreamer struct {
	stream decoder.Element
	cost   time.Duration
}

func (s *PipeLineStreamer) Name() string {
	return s.stream.Name()
}

func (s *PipeLineStreamer) Cost() int {
	return int(s.cost.Milliseconds())
}

type PipeLine struct {
	line         *speaker.Line
	buffer       *decoder.Samples
	format       *audio.Format
	wholeStreams []*PipeLineStreamer
	oneStreams   []*PipeLineStreamer

	locker sync.Mutex

	eleMixer     *element.Mixer
	eleVolume    *element.Volume
	eleResample  *element.Resample
	eleLineLM    *element.LineLevelMeter
	eleSpeakerLM *element.LineLevelMeter
	eleLineSPT   *element.LineLevelMeter
}

func (p *PipeLine) Len() int {
	return len(p.wholeStreams) + len(p.oneStreams)
}
func (p *PipeLine) Buffer() *decoder.Samples {
	return p.buffer
}

func (p *PipeLine) Prepend(s decoder.Element) {
	ps := []*PipeLineStreamer{{
		stream: s,
		cost:   0,
	}}

	if s.Type() == decoder.ET_OneSample {
		p.oneStreams = append(ps, p.wholeStreams...)
	} else if s.Type() == decoder.ET_WholeSamples {
		p.wholeStreams = append(ps, p.wholeStreams...)
	}
}

func (p *PipeLine) Add(s ...decoder.Element) {
	for _, ss := range s {
		ps := &PipeLineStreamer{
			stream: ss,
			cost:   0,
		}
		if ss.Type() == decoder.ET_OneSample {
			p.oneStreams = append(p.oneStreams, ps)
		} else if ss.Type() == decoder.ET_WholeSamples {
			p.wholeStreams = append(p.wholeStreams, ps)
		}
	}
}

func (p *PipeLine) Clear() {
	p.wholeStreams = p.wholeStreams[:0]
	p.oneStreams = p.oneStreams[:0]
}

func (p *PipeLine) Format() *audio.Format {
	return p.format
}

func (p *PipeLine) Close() error {
	for _, s := range p.wholeStreams {
		if sc, ok := s.stream.(decoder.StreamCloser); ok {
			sc.Close()
		}
	}
	for _, s := range p.oneStreams {
		if sc, ok := s.stream.(decoder.StreamCloser); ok {
			sc.Close()
		}
	}
	return nil
}

func (p *PipeLine) Stream() (int, error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	var t time.Time
	for _, s := range p.wholeStreams {
		t = time.Now()
		s.stream.Stream(p.buffer)
		s.cost = time.Since(t)
	}

	t = time.Now()
	for ch := 0; ch < p.buffer.Format.Layout.Count; ch++ {
		for i := 0; i < p.buffer.Size; i++ {
			for _, s := range p.oneStreams {
				s.stream.Sample(&p.buffer.Buffer[ch][i], ch, i)
			}
		}
	}
	for _, s := range p.oneStreams {
		s.cost = time.Since(t)
	}

	return p.buffer.LastSize, p.buffer.LastErr
}

func (p *PipeLine) Streamers() []*PipeLineStreamer {
	return append(p.oneStreams, p.wholeStreams...)
}

func (p *PipeLine) EleVolume() *element.Volume {
	return p.eleVolume
}

func (p *PipeLine) EleMixer() *element.Mixer {
	return p.eleMixer
}

func (p *PipeLine) EleResample() *element.Resample {
	return p.eleResample
}

func (p *PipeLine) EleLineLM() *element.LineLevelMeter {
	return p.eleLineLM
}

var pipeLineList = make([]*PipeLine, 0)

func Default() *PipeLine {
	return FromLine(speaker.DefaultLine())
}

func FromLine(line *speaker.Line) *PipeLine {
	for _, p := range pipeLineList {
		if p.line == line {
			return p
		}
	}
	return nil
}

func FromUUID(uuid string) *PipeLine {
	for _, p := range pipeLineList {
		if p.line.UUID == uuid {
			return p
		}
	}
	return Default()
}

func NewPipeLine(line *speaker.Line) *PipeLine {
	p := &PipeLine{
		line: line,
	}
	p.eleMixer = element.NewMixer()
	p.eleVolume = element.NewVolume(0.5)
	p.eleResample = element.NewResample(nil)
	p.eleLineLM = element.NewLineLevelMeter(line)

	p.Add(
		p.eleMixer,

		p.eleLineLM,
		p.eleVolume,

		p.eleResample,
	)

	pipeLineList = append(pipeLineList, p)
	return p
}
