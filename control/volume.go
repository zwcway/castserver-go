package control

import (
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

type Volume struct {
	f      Control
	Volume int // 0-100
	Mute   bool
}

func (s *Volume) Pack(id speaker.ID) (p *protocol.Package, err error) {
	s.f.cmd = Command_VOLUME
	s.f.spid = id

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
	if !sp.Config.AbsoluteVol {
		// 不支持绝对音量控制
		sp.Volume.SetVolume(vol)
		sp.Volume.SetMute(mute)
		return
	}
	s := Volume{
		Volume: int(vol * 100),
		Mute:   mute,
	}

	p, err := s.Pack(sp.Id)
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

}
