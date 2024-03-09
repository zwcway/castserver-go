package control

import (
	"time"

	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
)

type Time struct {
	f      Control
	server uint32
	offset uint16
}

func (s *Time) Pack() (p *protocol.Package, err error) {
	p, err = s.f.Pack()
	p.WriteUint32(s.server)
	p.WriteUint16(s.offset)

	err = nil
	return
}

func ControlTime(sp *speaker.Speaker) {
	t := &Time{
		f:      Control{Command_TIME, sp.ID},
		server: uint32(time.Now().UnixMilli()),
	}
	p, err := t.Pack()
	if err != nil {
		log.Error("encode time package error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(err))
		return
	}

	err = sp.WriteUDP(p.Bytes())
	if err != nil {
		log.Error("write speaker error", lg.Uint("speaker", uint64(sp.ID)), lg.Error(err))
		return
	}
}
