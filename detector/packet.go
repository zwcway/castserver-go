package detector

import (
	"net"
	"net/netip"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	utils "github.com/zwcway/castserver-go/common/utils"
)

type SpeakerResponse struct {
	Ver       uint8
	ID        speaker.SpeakerID
	Connected bool

	Addr netip.Addr
	MAC  net.HardwareAddr

	RateMask audio.AudioRateMask
	BitsMask audio.BitsMask

	DataPort uint16 // 用于接收 pcm 数据的端口，同样也是用于接收 控制帧 的端口,尽可能节省设备端资源

	AbsoluteVol bool // 是否支持音量控制
	PowerSave   bool // 是否支持开关机/低电量
}

func byte2bool(b byte) bool {
	return b != 0
}
func bool2byte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func (r *SpeakerResponse) Unpack(p *protocol.Package) (err error) {
	var (
		i8  uint8
		i16 uint16
		i32 uint32
		bs  []byte
	)
	i8, err = p.ReadUint8()
	if err != nil {
		err = newUnpackError("protocol type", p.LastBytes(1), err)
		return
	}
	if protocol.PT_SpeakerInfo != protocol.Type(i8) {
		err = newUnpackError("protocol type error", p.LastBytes(1), nil)
		return
	}

	i8, err = p.ReadUint8()
	if err != nil {
		err = newUnpackError("version", p.LastBytes(1), err)
		return
	}
	r.Ver = i8 >> 4
	r.Connected = byte2bool((i8 >> 3) & 0x01)
	is6 := byte2bool((i8 >> 2) & 0x01)

	if is6 {
		bs, err = p.Read(net.IPv6len)
		if err != nil {
			err = newUnpackError("ipv6", bs, err)
			return
		}
		r.Addr = netip.AddrFrom16([16]byte(bs))
	} else {
		bs, err = p.Read(net.IPv4len)
		if err != nil {
			err = newUnpackError("ipv4", bs, err)
			return
		}
		r.Addr = netip.AddrFrom4([4]byte(bs))
	}

	i32, err = p.ReadUint32()
	if err != nil {
		err = newUnpackError("speaker id", p.LastBytes(4), err)
		return
	}
	r.ID = speaker.SpeakerID(i32)

	bs, err = p.Read(6)
	if err != nil {
		err = newUnpackError("mac address", bs, err)
		return
	}
	r.MAC = bs
	if !utils.MacIsValid(r.MAC) {
		err = newUnpackError("mac address", bs, nil)
		return
	}

	i16, err = p.ReadUint16()
	if err != nil {
		err = newUnpackError("rate mask", p.LastBytes(2), err)
		return
	}
	r.RateMask = audio.AudioRateMask(i16)
	if !r.RateMask.IsValid() {
		err = newUnpackError("rate mask", p.LastBytes(2), err)
	}

	i16, err = p.ReadUint16()
	if err != nil {
		err = newUnpackError("bits mask", p.LastBytes(2), err)
		return
	}
	r.BitsMask = audio.BitsMask(i16)
	if !r.BitsMask.IsValid() {
		err = newUnpackError("bits mask", p.LastBytes(2), err)
		return
	}

	i16, err = p.ReadUint16()
	if err != nil {
		err = newUnpackError("data port", p.LastBytes(2), err)
		return
	}
	r.DataPort = i16
	if !utils.PortIsValid(r.DataPort) {
		err = newUnpackError("data port", p.LastBytes(2), err)
		return
	}

	i32, err = p.ReadUint32()
	if err != nil {
		err = newUnpackError("extern error", p.LastBytes(2), err)
		return
	}
	r.AbsoluteVol = ((i32 >> 0) & 0x01) == 1
	r.PowerSave = ((i32 >> 1) & 0x01) == 1

	return
}

type ServerType int

const (
	ST_Start ServerType = 1 + iota
	ST_Response
	ST_Exit
)

type ServerResponse struct {
	Ver  uint8
	Type ServerType
	Addr netip.Addr
	Port uint16
}

func (sr *ServerResponse) Pack() (p *protocol.Package, err error) {
	p = protocol.NewPackage(32)
	err = p.WriteUint8(uint8(protocol.PT_ServerInfo))
	if err != nil {
		return
	}

	is6 := sr.Addr.Is6()

	i8 := sr.Ver<<4 | ((uint8(sr.Type) << 1) & 0x0E) | bool2byte(is6)
	err = p.WriteUint8(i8)
	if err != nil {
		return
	}

	if is6 {
		ipv6 := sr.Addr.As16()
		err = p.Write(ipv6[:])
		if err != nil {
			return
		}
	} else {
		ipv4 := sr.Addr.As4()
		err = p.Write(ipv4[:])
		if err != nil {
			return
		}
	}
	err = p.WriteUint16(sr.Port)
	return
}
