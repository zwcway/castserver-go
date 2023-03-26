package control

import (
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
)

type Command uint32

const (
	Command_UNKNOWN Command = iota
	Command_SAMPLE
	Command_CHUNK
	Command_TIME
	Command_VOLUME

	Command_MAX
)

type Control struct {
	cmd  Command
	spid speaker.SpeakerID
}

func (f *Control) Pack() (p *protocol.Package, err error) {
	p = protocol.NewPackage(16)
	p.WriteUint8(uint8(protocol.PT_Control))
	p.WriteUint8(uint8((protocol.VERSION&0x0F)<<4) | uint8(f.cmd&0x0F))
	p.WriteUint32(uint32(f.spid))
	return
}
