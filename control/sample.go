package control

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

type Sample struct {
	f       Control
	bit     audio.Bits
	rate    audio.Rate
	channel audio.Channel
}

func (s *Sample) Pack() (p *protocol.Package, err error) {
	p, _ = s.f.Pack()
	p.WriteUint8(uint8(s.bit) << 4 & uint8(s.rate))
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
		log.Error("encode sample package error", zap.String("speaker", sp.String()), zap.Error(err))
		return
	}

	err = sp.WriteUDP(p.Bytes())
	if err != nil {
		log.Error("ControlSample error", zap.String("speaker", sp.String()), zap.Error(err))
		return
	}
}
