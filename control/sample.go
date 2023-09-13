package control

import (
	"github.com/zwcway/castserver-go/common/audio"
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
)

type Sample struct {
	f       Control
	bit     audio.Bits
	rate    audio.Rate
	channel audio.Channel
}

func (s *Sample) Pack() (p *protocol.Package, err error) {
	p, _ = s.f.Pack()
	p.WriteUint8((uint8(s.bit) << 4) | uint8(s.rate))
	p.WriteUint8(uint8(s.channel))

	return
}

func ControlSample(sp *speaker.Speaker) {
	s := Sample{
		f:       Control{Command_SAMPLE, sp.ID},
		bit:     sp.SampleBits(),
		rate:    sp.SampleRate(),
		channel: sp.SampleChannel(),
	}

	p, err := s.Pack()
	if err != nil {
		log.Error("encode sample package error", log1.String("speaker", sp.String()), log1.Error(err))
		return
	}

	err = sp.WriteUDP(p.Bytes())
	if err != nil {
		log.Error("ControlSample error", log1.String("speaker", sp.String()), log1.Error(err))
		return
	}
}
