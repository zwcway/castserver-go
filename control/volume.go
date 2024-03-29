package control

import (
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
)

type Volume struct {
	f      Control
	Volume int // 0-100
	Mute   bool
}

func (s *Volume) Pack(id speaker.SpeakerID) (p *protocol.Package, err error) {
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
		sp.VolumeEle.SetVolume(vol)
		sp.VolumeEle.SetMute(mute)
		return
	}
	s := Volume{
		Volume: int(vol * 100),
		Mute:   mute,
	}

	p, err := s.Pack(sp.ID)
	if err != nil {
		log.Error("encode volume package error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(err))
		return
	}

	err = sp.WriteUDP(p.Bytes())
	if err != nil {
		log.Error("write speaker error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(err))
		return
	}
}
