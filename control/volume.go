package control

import (
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder/pipeline"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type Volume struct {
	f      Control
	Volume int
	Mute   bool
}

func (s *Volume) Pack() (p *protocol.Package, err error) {
	p, _ = s.f.Pack()
	p.WriteUint8(uint8(s.Volume))
	if s.Mute {
		p.WriteUint8(1)
	} else {
		p.WriteUint8(0)
	}

	return
}

func ControlSpeakerVolume(sp *speaker.Speaker) {
	if !sp.AbsoluteVol {
		return
	}
	s := Volume{
		Volume: sp.Volume,
		Mute:   sp.IsMute,
	}

	p, err := s.Pack()
	if err != nil {
		log.Error("encode volume package error", zap.Uint32("speaker", uint32(sp.ID)), zap.Error(err))
		return
	}

	err = sp.WriteUDP(p.Bytes())
	if err != nil {
		log.Error("write speaker error", zap.Uint32("speaker", uint32(sp.ID)), zap.Error(err))
		return
	}
}

func ControlLineVolume(line *speaker.Line, vol float64, mute bool) {
	p := pipeline.FromLine(line)
	if p == nil || p.EleVolume() == nil {
		return
	}
	p.EleVolume().SetVolume(vol)
	p.EleVolume().SetMute(mute)

	websockets.BroadcastLineEvent(line, websockets.Event_Line_Edited)
}
