package control

import (
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type Volume struct {
	f      Control
	Volume int // 0-100
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

func ControlSpeakerVolume(sp *speaker.Speaker, vol float64, mute bool) {
	if !sp.AbsoluteVol {
		// 不支持绝对音量控制
		sp.Volume.SetVolume(vol)
		sp.Volume.SetMute(mute)
		return
	}
	s := Volume{
		Volume: int(vol * 100),
		Mute:   mute,
	}

	p, err := s.Pack()
	if err != nil {
		log.Error("encode volume package error", zap.Uint32("speaker", uint32(sp.Id)), zap.Error(err))
		return
	}

	err = sp.WriteUDP(p.Bytes())
	if err != nil {
		log.Error("write speaker error", zap.Uint32("speaker", uint32(sp.Id)), zap.Error(err))
		return
	}
}

func ControlLineVolume(line *speaker.Line, vol float64, mute bool) {
	p := line.Input.PipeLine
	if p == nil || line.Volume == nil {
		return
	}
	line.Volume.SetVolume(vol)
	line.Volume.SetMute(mute)

	websockets.BroadcastLineEvent(line, websockets.Event_Line_Edited)
}
