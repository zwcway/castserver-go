package pusher

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/protocol"
)

type ServerPush struct {
	Ver      uint8
	Compress uint8
	Rate     audio.Rate
	Bits     audio.Bits
	Time     uint16
	Samples  []byte
}

const ServerPushHeaderSize uint16 = 7

var globalBuffer = protocol.NewPackage(65535)

func (s *ServerPush) Pack() (p *protocol.Package, err error) {
	p = globalBuffer
	p.Reset()

	err = p.WriteUint8(uint8(protocol.PT_SpeakerDataPush))
	if err != nil {
		return
	}
	err = p.WriteUint8((s.Ver << 4) | s.Compress)
	if err != nil {
		return
	}
	err = p.WriteUint8((uint8(s.Bits) << 4) | uint8(s.Rate))
	if err != nil {
		return
	}

	err = p.WriteUint16(s.Time)
	if err != nil {
		return
	}

	// len
	err = p.WriteUint16(uint16(len(s.Samples)))
	if err != nil {
		return
	}

	err = p.Write(s.Samples)
	return
}
