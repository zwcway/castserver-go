package pusher

import "github.com/zwcway/castserver-go/common/protocol"

type ServerPush struct {
	Ver     uint8
	Seq     uint32
	Time    uint32
	Samples []byte
}

var globalBuffer = protocol.NewPackage(1024)

func (s *ServerPush) Pack() (p *protocol.Package, err error) {
	p = globalBuffer
	p.Reset()

	err = p.WriteUint8(uint8(protocol.PT_SpeakerDataPush))
	if err != nil {
		return
	}
	err = p.WriteUint32(s.Seq)

	return
}
