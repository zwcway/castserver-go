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
		f:       Control{Command_SAMPLE, sp.Id},
		bit:     sp.Bits,
		rate:    sp.Rate,
		channel: sp.Channel,
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

func ControlRate(sp *speaker.Speaker, rate audio.Rate) {
	sp.Rate = rate
	ControlSample(sp)
}

func ControlBits(sp *speaker.Speaker, bits audio.Bits) {
	sp.Bits = bits
	ControlSample(sp)
}
func ControlChannel(sp *speaker.Speaker, ch audio.Channel) {
	if sp.Channel == ch {
		return
	}

	sp.ChangeChannel(ch)
	ControlSample(sp)
}
