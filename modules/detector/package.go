package detector

import (
	"encoding/binary"
	"net"
	"net/netip"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	utils "github.com/zwcway/castserver-go/utils"
)

type Response struct {
	Ver       uint8
	ID        speaker.SpeakerID
	Connected bool

	Addr netip.Addr
	MAC  net.HardwareAddr

	RateMask audio.AudioRateMask
	BitsMask audio.AudioBitsMask

	DataPort uint16
	CastPort uint16
}

func byte2bool(b byte) bool {
	return b == 1
}

func packageSize(is6 bool) int {
	if is6 {
		return 35
	}

	return 23
}

func unPack(p []byte) (*Response, error) {
	var r Response
	var err error
	var i int = 0

	r.Ver = p[i] >> 4
	r.Connected = byte2bool((p[i] >> 3) & 0x01)
	is6 := byte2bool((p[i] >> 2) & 0x01)
	i += 1

	if is6 {
		r.Addr, err = netip.ParseAddr(string(p[i : i+16 : 16]))
		if err != nil {
			return nil, &ResponseDecodeError{"Addr", p[i : i+16 : 16], err}
		}
		i += 16
	} else {
		r.Addr, err = netip.ParseAddr(string(p[i : i+4 : 4]))
		if err != nil {
			return nil, &ResponseDecodeError{"Addr", p[i : i+4 : 4], err}
		}
		i += 4
	}

	r.ID = speaker.SpeakerID(binary.BigEndian.Uint32(p[i : i+4 : 4]))
	i += 4

	r.MAC = p[i : i+6]
	if !utils.MacIsValid(r.MAC) {
		return nil, &ResponseDecodeError{"RateMask", r.MAC, nil}
	}
	i += 6

	r.RateMask = audio.AudioRateMask(binary.BigEndian.Uint16(p[i : i+2 : 2]))
	if !r.RateMask.IsValid() {
		return nil, &ResponseDecodeError{"RateMask", p[i : i+2 : 2], nil}
	}
	i += 2

	r.BitsMask = audio.AudioBitsMask(binary.BigEndian.Uint16(p[i : i+2 : 2]))
	if !r.BitsMask.IsValid() {
		return nil, &ResponseDecodeError{"BitsMask", p[i : i+2 : 2], nil}
	}
	i += 2

	r.DataPort = binary.BigEndian.Uint16(p[i : i+2 : 2])
	if !utils.PortIsValid(r.DataPort) {
		return nil, &ResponseDecodeError{"DataPort", p[i : i+2 : 2], nil}
	}
	i += 2

	r.CastPort = binary.BigEndian.Uint16(p[i : i+2 : 2])
	if !utils.PortIsValid(r.CastPort) {
		return nil, &ResponseDecodeError{"CastPort", p[i : i+2 : 2], nil}
	}
	i += 2

	return &r, nil
}
